package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api/executor"
	"github.com/axsh/openvdc/cmd"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/golang/protobuf/proto"
	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
	mesosutil "github.com/mesos/mesos-go/mesosutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

type VDCExecutor struct {
	hypervisorProvider hypervisor.HypervisorProvider
	ctx                context.Context
	nodeInfo           *model.ExecutorNode
}

func newVDCExecutor(ctx context.Context, provider hypervisor.HypervisorProvider, node *model.ExecutorNode) *VDCExecutor {
	return &VDCExecutor{
		hypervisorProvider: provider,
		ctx:                ctx,
		nodeInfo:           node,
	}
}

func (exec *VDCExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	log.Infoln("Registered Executor on slave ", slaveInfo.GetHostname())
	exec.nodeInfo.Id = slaveInfo.GetId().GetValue()
	err := model.Cluster(exec.ctx).Register(exec.nodeInfo)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infoln("Registered on OpenVDC cluster service: ", exec.nodeInfo)
}

func (exec *VDCExecutor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.SlaveInfo) {
	log.Infoln("Re-registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *VDCExecutor) Disconnected(driver exec.ExecutorDriver) {
	log.Infoln("Executor disconnected.")
}

func (exec *VDCExecutor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	log.Infoln("Launching task", taskInfo.GetName(), "with command", taskInfo.Command.GetValue())

	_, err := driver.SendStatusUpdate(&mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_STARTING.Enum(),
	})
	if err != nil {
		log.WithError(err).Errorln("Couldn't send status update")
		return
	}
	if err := exec.bootInstance(driver, taskInfo); err != nil {
		return
	}

	_, err = driver.SendStatusUpdate(&mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	})
	if err != nil {
		log.WithError(err).Errorln("Couldn't send status update")
		return
	}
}

func recordFailedState(ctx context.Context, driver exec.ExecutorDriver, instanceID string, failureType model.FailureMessage_ErrorType, lastErr error) error {
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"error_type":  failureType.String(),
		"last_error":  lastErr,
	})
	err1 := model.Instances(ctx).AddFailureMessage(instanceID, failureType)
	if err1 != nil {
		log.WithError(err1).Errorln("Failed Instances.AddFailureMessage")
	}
	status := &mesos.TaskStatus{
		TaskId: mesosutil.NewTaskID(instanceID),
		State:  mesos.TaskState_TASK_FAILED.Enum(),
	}
	if lastErr != nil {
		status.Message = proto.String(lastErr.Error())
	}
	_, err2 := driver.SendStatusUpdate(status)
	if err2 != nil {
		log.WithError(err2).Error("Failed to SendStatusUpdate TASK_FAILED")
	}
	if err1 == nil && err2 == nil {
		log.Info("Proceeded recording task failure")
	}
	return nil
}

func (exec *VDCExecutor) bootInstance(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) error {
	instanceID := taskInfo.GetTaskId().GetValue()
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	inst, err := model.Instances(ctx).FindByID(instanceID)
	if err != nil {
		return errors.Wrap(err, "Failed model.Instance.FindByID")
	}
	// Assert the race for scheduling slave and instance.
	if inst.GetSlaveId() != taskInfo.GetSlaveId().GetValue() {
		log.Fatalf("BUGON: Found mismatch for SlaveID assignment between instance and Mesos message: instance expects %s but Mesos says %s",
			inst.GetSlaveId(),
			taskInfo.GetSlaveId().String())
	}

	// Apply FAILED terminal state in case of error.
	finState := model.InstanceState_FAILED
	var lastErr error
	defer func() {
		if err := model.Instances(ctx).UpdateState(instanceID, finState); err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		if finState == model.InstanceState_FAILED {
			recordFailedState(ctx, driver, instanceID, model.FailureMessage_FAILED_BOOT, lastErr)
		}
		log.WithField("fin_state", finState).Info("Proceeded defer func() at bootInstance()")
		model.Close(ctx)
	}()

	log = log.WithField("state", model.InstanceState_STARTING.String())
	if lastErr = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STARTING); lastErr != nil {
		log.WithError(lastErr).Error("Failed Instances.UpdateState")
		return lastErr
	}

	// Reload instance object from datastore.x
	inst, lastErr = model.Instances(ctx).FindByID(instanceID)
	if lastErr != nil {
		log.WithError(lastErr).Error("Failed Instances.FindyByID")
		return lastErr
	}

	hv, lastErr := exec.hypervisorProvider.CreateDriver(inst, inst.ResourceTemplate())
	if lastErr != nil {
		return lastErr
	}

	log.Infof("Creating instance")
	if lastErr = hv.CreateInstance(); lastErr != nil {
		log.WithError(lastErr).Error("Failed CreateInstance")
		return lastErr
	}
	log.Infof("Starting instance")
	if lastErr = hv.StartInstance(); lastErr != nil {
		log.WithError(lastErr).Error("Failed StartInstance")
		return lastErr
	}
	log.Infof("Instance launched successfully")
	// Here can bring the instance state to RUNNING finally.
	finState = model.InstanceState_RUNNING
	return nil
}

func (exec *VDCExecutor) startInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Apply FAILED terminal state in case of error.
	finState := model.InstanceState_FAILED
	var lastErr error
	defer func() {
		if finState == model.InstanceState_FAILED {
			recordFailedState(ctx, driver, instanceID, model.FailureMessage_FAILED_START, lastErr)
		}
		if err := model.Instances(ctx).UpdateState(instanceID, finState); err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	inst, lastErr := model.Instances(ctx).FindByID(instanceID)
	if lastErr != nil {
		log.WithError(lastErr).Error("Failed Instances.FindyByID")
		return lastErr
	}

	hv, lastErr := exec.hypervisorProvider.CreateDriver(inst, inst.ResourceTemplate())
	if lastErr != nil {
		return lastErr
	}

	if lastErr = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STARTING); lastErr != nil {
		log.WithError(lastErr).WithField("state", model.InstanceState_STOPPING).Error("Failed Instances.UpdateState")
		return lastErr
	}

	log.Infof("Starting instance")
	if lastErr = hv.StartInstance(); lastErr != nil {
		log.WithError(lastErr).Error("Failed StartInstance")
		return lastErr
	}
	log.Infof("Instance started successfully")
	// Here can bring the instance state to RUNNING finally.
	finState = model.InstanceState_RUNNING
	return nil
}

func (exec *VDCExecutor) stopInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Apply FAILED terminal state in case of error.
	finState := model.InstanceState_FAILED
	var lastErr error
	defer func() {
		if finState == model.InstanceState_FAILED {
			recordFailedState(ctx, driver, instanceID, model.FailureMessage_FAILED_STOP, lastErr)
		}
		if err := model.Instances(ctx).UpdateState(instanceID, finState); err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	inst, lastErr := model.Instances(ctx).FindByID(instanceID)
	if lastErr != nil {
		log.WithError(lastErr).Error("Failed Instances.FindyByID")
		return lastErr
	}

	hv, lastErr := exec.hypervisorProvider.CreateDriver(inst, inst.ResourceTemplate())
	if lastErr != nil {
		return lastErr
	}

	if lastErr = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STOPPING); lastErr != nil {
		log.WithError(lastErr).WithField("state", model.InstanceState_STOPPING).Error("Failed Instances.UpdateState")
		return lastErr
	}

	log.Infof("Stopping instance")
	if lastErr = hv.StopInstance(); lastErr != nil {
		log.WithError(lastErr).Error("Failed StopInstance")
		return lastErr
	}
	log.Infof("Instance stopped successfully")
	// Here can bring the instance state to STOPPED finally.
	finState = model.InstanceState_STOPPED
	return nil
}

func (exec *VDCExecutor) rebootInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Apply FAILED terminal state in case of error.
	finState := model.InstanceState_FAILED
	var lastErr error
	defer func() {
		if finState == model.InstanceState_FAILED {
			recordFailedState(ctx, driver, instanceID, model.FailureMessage_FAILED_REBOOT, lastErr)
		}
		if err := model.Instances(ctx).UpdateState(instanceID, finState); err != nil {
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	inst, lastErr := model.Instances(ctx).FindByID(instanceID)
	if lastErr != nil {
		log.WithError(lastErr).Error("Failed Instances.FindyByID")
		return lastErr
	}

	// .LastState must be set to REBOOTING at the API server.
	if inst.GetLastState().GetState() != model.InstanceState_REBOOTING {
		lastErr = errors.Errorf("Invalid instance state for reboot operation: %s", inst.GetLastState().GetState())
		return lastErr
	}

	hv, lastErr := exec.hypervisorProvider.CreateDriver(inst, inst.ResourceTemplate())
	if lastErr != nil {
		return lastErr
	}

	log.Infof("Rebooting instance")
	if lastErr = hv.RebootInstance(); lastErr != nil {
		log.Error("Failed RebootInstance")
		return lastErr
	}

	log.Infof("Instance rebooted successfully")
	finState = model.InstanceState_RUNNING

	return nil
}

func (exec *VDCExecutor) terminateInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(log.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Apply FAILED terminal state in case of error.
	finState := model.InstanceState_FAILED
	var lastErr error
	defer func() {
		if finState == model.InstanceState_FAILED {
			recordFailedState(ctx, driver, instanceID, model.FailureMessage_FAILED_TERMINATE, lastErr)
		}
		if err := model.Instances(ctx).UpdateState(instanceID, finState); err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	inst, lastErr := model.Instances(ctx).FindByID(instanceID)
	if lastErr != nil {
		return errors.Wrap(lastErr, "Failed instances.FindByID")
	}

	originalState := inst.GetLastState().GetState()

	hv, lastErr := exec.hypervisorProvider.CreateDriver(inst, inst.ResourceTemplate())
	if lastErr != nil {
		return errors.Wrap(lastErr, "Failed hypervisorProvider.CreateDriver")
	}

	if lastErr = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_SHUTTINGDOWN); lastErr != nil {
		log.WithError(lastErr).WithField("state", model.InstanceState_SHUTTINGDOWN).Error("Failed Instances.UpdateState")
		return lastErr
	}

	// Trying to stop an already stopped container results in an error
	// causing the container to not get destroyed.

	if originalState != model.InstanceState_STOPPED {
		log.Infof("Shuttingdown instance")
		if lastErr = hv.StopInstance(); lastErr != nil {
			log.WithError(lastErr).Error("Failed StopInstance")
			return lastErr
		}
	}

	if lastErr = hv.DestroyInstance(); lastErr != nil {
		log.WithError(lastErr).Error("Failed DestroyInstance")
		return lastErr
	}
	log.Infof("Instance terminated successfully")
	// Here can bring the instance state to TERMINATED finally.
	finState = model.InstanceState_TERMINATED
	return nil
}

func (exec *VDCExecutor) KillTask(driver exec.ExecutorDriver, taskID *mesos.TaskID) {
	log.Infoln("Kill task")
}

func (exec *VDCExecutor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {

	parts := strings.Split(msg, "_")
	command := parts[0]
	taskId := mesosutil.NewTaskID(parts[1])

	log.Infoln("--------------FrameworkMessage---------------")
	log.Infoln("command: ", command)
	log.Infoln("taskId: ", taskId)
	log.Infoln("---------------------------------------------")

	switch command {
	case "start":
		if err := exec.startInstance(driver, taskId.GetValue()); err != nil {
			log.WithError(err).Error("Failed to start instance")
		}
	case "stop":
		if err := exec.stopInstance(driver, taskId.GetValue()); err != nil {
			log.WithError(err).Error("Failed to stop instance")
		}
	case "reboot":
		if err := exec.rebootInstance(driver, taskId.GetValue()); err != nil {
			log.WithError(err).Error("Failed to reboot instance")
		}
	case "destroy":
		if err := exec.terminateInstance(driver, taskId.GetValue()); err != nil {
			log.WithError(err).Error("Failed to terminate instance")
			// driver.SendStatusUpdate() with TASK_FAILED message is sent in terminateInstance()
			break
		}
		_, err := driver.SendStatusUpdate(&mesos.TaskStatus{
			TaskId: taskId,
			State:  mesos.TaskState_TASK_FINISHED.Enum(),
			Source: mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
		})
		if err != nil {
			log.WithError(errors.WithStack(err)).Error("Couldn't send status update")
		}
	default:
		log.WithField("msg", msg).Errorln("FrameworkMessage unrecognized.")
	}
}

func (exec *VDCExecutor) Shutdown(driver exec.ExecutorDriver) {
	log.Infoln("Shutting down the executor")
}

func (exec *VDCExecutor) Error(driver exec.ExecutorDriver, err string) {
	log.Errorln("Got error message:", err)
}

var rootCmd = &cobra.Command{
	Use:   "openvdc-executor",
	Short: "",
	Long:  ``,
	Run:   execute,
}

var DefaultConfPath string
var zkAddr backend.ZkEndpoint

const defaultExecutorAPIPort = "19372"
const defaultSSHPort = "29876"

func startExecutorAPIServer(ctx context.Context, listener net.Listener) *executor.ExecutorAPIServer {
	s := executor.NewExecutorAPIServer(&zkAddr, ctx)
	go s.Serve(listener)
	return s
}

func init() {
	viper.SetDefault("hypervisor.driver", "null")
	viper.SetDefault("zookeeper.endpoint", "zk://localhost/openvdc")
	viper.SetDefault("executor-api.listen", "0.0.0.0:"+defaultExecutorAPIPort)
	viper.SetDefault("executor-api.advertise-ip", "")
	viper.SetDefault("console.ssh.listen", "0.0.0.0:"+defaultSSHPort)
	viper.SetDefault("console.ssh.advertise-ip", "")

	cobra.OnInitialize(initConfig)
	pfs := rootCmd.PersistentFlags()
	pfs.String("config", DefaultConfPath, "Load config file from the path")
	pfs.String("hypervisor", viper.GetString("hypervisor.driver"), "Hypervisor driver name")
	viper.BindPFlag("hypervisor.driver", pfs.Lookup("hypervisor"))
	pfs.String("zk", viper.GetString("zookeeper.endpoint"), "Zookeeper address")
	viper.BindPFlag("zookeeper.endpoint", pfs.Lookup("zk"))
}

func initConfig() {
	f := rootCmd.PersistentFlags().Lookup("config")
	if f.Changed {
		viper.SetConfigFile(f.Value.String())
	} else {
		viper.SetConfigFile(DefaultConfPath)
		viper.SetConfigType("toml")
	}
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if viper.ConfigFileUsed() == DefaultConfPath && os.IsNotExist(err) {
			// Ignore default conf file does not exist.
			return
		}
		log.WithError(err).Fatalf("Failed to load config %s", viper.ConfigFileUsed())
	}
}

func init() {
	// Initialize golang/glog flags used by mesos-go.
	flag.CommandLine.Parse([]string{})
	flag.Set("logtostderr", "true")
}

func execute(cmd *cobra.Command, args []string) {
	err := zkAddr.Set(viper.GetString("zookeeper.endpoint"))
	if err != nil {
		log.WithError(err).Fatal("Invalid zookeeper endpoint: ", viper.GetString("zookeeper.endpoint"))
	}

	provider, ok := hypervisor.FindProvider(viper.GetString("hypervisor.driver"))
	if ok == false {
		log.Fatalln("Unknown hypervisor name:", viper.GetString("hypervisor.driver"))
	}
	if err := provider.LoadConfig(viper.GetViper()); err != nil {
		log.WithError(err).Fatal("Failed to apply hypervisor configuration")
	}
	log.Infof("Initializing executor: hypervisor %s\n", provider.Name())

	ctx, err := model.ClusterConnect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connect to cluster service %s", zkAddr.String())
	}
	defer func() {
		err := model.ClusterClose(ctx)
		if err != nil {
			log.WithError(err).Error("Failed ClusterClose")
		}
	}()
	ctx, err = model.Connect(ctx, &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connecto to model server: %s", zkAddr.String())
	}
	defer func() {
		if err := model.Close(ctx); err != nil {
			log.WithError(err).Error("Failed model.Close")
		}
	}()

	executorAPIListener, err := net.Listen("tcp", viper.GetString("executor-api.listen"))
	if err != nil {
		log.WithError(err).Fatalln("Faild to bind address for Executor gRPC API: ", viper.GetString("executor-api.listen"))
	}
	s := startExecutorAPIServer(ctx, executorAPIListener)
	defer s.GracefulStop()
	log.Infof("Listening Executor gRPC API on %s", executorAPIListener.Addr().String())
	exposedExecutorAPIAddr := executorAPIListener.Addr().String()
	if viper.GetString("executor-api.advertise-ip") != "" {
		_, port, err := net.SplitHostPort(exposedExecutorAPIAddr)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse host:port: ", exposedExecutorAPIAddr)
		}
		exposedExecutorAPIAddr = net.JoinHostPort(viper.GetString("executor-api.advertise-ip"), port)
		log.Infof("Exposed Executor gRPC API on %s", exposedExecutorAPIAddr)
	}

	sshListener, err := net.Listen("tcp", viper.GetString("console.ssh.listen"))
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen SSH on %s", sshListener.Addr().String())
	}
	defer sshListener.Close()

	sshd := NewSSHServer(provider, ctx)
	if err := sshd.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to setup SSH Server")
	}
	go sshd.Run(sshListener)
	log.Infof("Listening SSH on %s", sshListener.Addr().String())
	exposedSSHAddr := sshListener.Addr().String()
	if viper.GetString("console.ssh.advertise-ip") != "" {
		_, port, err := net.SplitHostPort(exposedSSHAddr)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse host:port: ", exposedSSHAddr)
		}
		exposedSSHAddr = net.JoinHostPort(viper.GetString("console.ssh.advertise-ip"), port)
		log.Infof("Exposed SSH on %s", exposedSSHAddr)
	}

	node := &model.ExecutorNode{
		GrpcAddr: exposedExecutorAPIAddr,
		Console: &model.Console{
			Type:     model.Console_SSH,
			BindAddr: exposedSSHAddr,
		},
	}
	dconfig := exec.DriverConfig{
		Executor: newVDCExecutor(ctx, provider, node),
	}
	driver, err := exec.NewMesosExecutorDriver(dconfig)
	if err != nil {
		log.WithError(err).Fatal("Couldn't create ExecutorDriver")
	}

	_, err = driver.Start()
	if err != nil {
		log.WithError(err).Fatalln("ExecutorDriver wasn't able to start")
	}
	log.Infoln("Process running")

	_, err = driver.Join()
	if err != nil {
		log.WithError(err).Fatalln("Something went wrong with the driver")
	}
	log.Infoln("Executor shutting down")
}

func main() {
	log.SetFormatter(&cmd.LogFormatter{})
	log.SetLevel(log.DebugLevel)
	rootCmd.AddCommand(cmd.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
