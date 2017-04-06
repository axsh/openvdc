package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/api/executor"
	"github.com/axsh/openvdc/cmd"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
	mesosutil "github.com/mesos/mesos-go/mesosutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var log = logrus.WithField("context", "vdc-executor")

type VDCExecutor struct {
	hypervisorProvider hypervisor.HypervisorProvider
	ctx                context.Context
	nodeInfo           *model.ExecutorNode
}

func newVDCExecutor(ctx context.Context, node *model.ExecutorNode) *VDCExecutor {
	return &VDCExecutor{
		ctx:      ctx,
		nodeInfo: node,
	}
}

func (exec *VDCExecutor) GetHypervisorProvider() hypervisor.HypervisorProvider {
	return exec.hypervisorProvider
}

func (exec *VDCExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	log.Infoln("Registering Executor on slave ", slaveInfo.GetHostname())
	// Read and validate passed slave attributes: /etc/mesos-slave/attributes or --attributes
	for _, attr := range slaveInfo.Attributes {
		switch attr.GetName() {
		case "hypervisor":
			var ok bool
			exec.hypervisorProvider, ok = hypervisor.FindProvider(attr.GetText().GetValue())
			if !ok {
				log.Errorf("Unknown hypervisor driver: %s", attr.GetText().GetValue())
			}
		}
	}
	if exec.hypervisorProvider == nil {
		_, err := driver.Abort()
		log.WithError(err).Error("Failed to find 'hypervisor' attribute")
		return
	}
	if err := exec.hypervisorProvider.LoadConfig(viper.GetViper()); err != nil {
		log.WithError(err).Error("Failed hypervisorProvider.LoadConfig")
		driver.Abort()
		return
	}

	exec.nodeInfo.Id = slaveInfo.GetId().GetValue()
	err := model.Cluster(exec.ctx).Register(exec.nodeInfo)
	if err != nil {
		driver.Abort()
		log.WithError(err).Error("Failed OpenVDC cluster registration")
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

	runStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	}
	_, err := driver.SendStatusUpdate(runStatus)
	if err != nil {
		log.WithError(err).Error("Couldn't send status update")
	}

	instanceID := taskInfo.GetTaskId().GetValue()

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
	}
	instance, err := model.Instances(ctx).FindByID(instanceID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch instance %s", instanceID)
	}
	instanceState := instance.GetLastState()

	if instanceState.State == model.InstanceState_QUEUED {
		err = exec.bootInstance(driver, taskInfo)
		if err != nil {
			_, err := driver.SendStatusUpdate(&mesos.TaskStatus{
				TaskId: taskInfo.GetTaskId(),
				State:  mesos.TaskState_TASK_FAILED.Enum(),
			})
			if err != nil {
				log.WithError(err).Error("Failed to SendStatusUpdate TASK_FAILED")
			}
			if err := exec.Failure(taskInfo.GetTaskId().GetValue(), model.FailureMessage_FAILED_BOOT); err != nil {
				log.WithError(err).Errorf("Failed to record failure message: %s", model.FailureMessage_FAILED_BOOT.String())
			}
		}
	} else {
		if err := exec.recoverInstance(taskInfo.GetTaskId().GetValue(), *instanceState); err != nil {
			log.WithError(err).Error("Failed recoverInstance")
		}
	}
}

func (exec *VDCExecutor) recoverInstance(instanceID string, instanceState model.InstanceState) error {
	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return errors.Wrapf(err, "Hypervisorprovider failed to create driver. InstanceID:  %s", instanceID)
	}
	err = hv.Recover(instanceState)
	if err != nil {
		return errors.Wrapf(err, "Hypervisor failed to recover instance. InstanceID: %s", instanceID)
	}
	return nil
}

func (exec *VDCExecutor) bootInstance(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) error {
	instanceID := taskInfo.GetTaskId().GetValue()
	log := log.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Push back to the initial state in case of error.
	finState := model.InstanceState_REGISTERED
	defer func() {
		err = model.Instances(ctx).UpdateState(instanceID, finState)
		if err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	log = log.WithField("state", model.InstanceState_STARTING.String())
	err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STARTING)
	if err != nil {
		log.WithError(err).Error("Failed Instances.UpdateState")
		return err
	}

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return err
	}

	inst, err := model.Instances(ctx).FindByID(instanceID)
	if err != nil {
		log.WithError(err).Error("Failed Instances.FindyByID")
		return err
	}

	log.Infof("Creating instance")
	err = hv.CreateInstance(inst, inst.ResourceTemplate())
	if err != nil {
		log.WithError(err).Error("Failed CreateInstance")
		return err
	}
	log.Infof("Starting instance")
	err = hv.StartInstance()
	if err != nil {
		log.WithError(err).Error("Failed StartInstance")
		return err
	}
	log.Infof("Instance launched successfully")
	// Here can bring the instance state to RUNNING finally.
	finState = model.InstanceState_RUNNING

	err = model.Instances(ctx).UpdateConnectionStatus(instanceID, model.ConnectionStatus_CONNECTED)

	if err != nil {
		return errors.Wrapf(err, "Couldn't update instance connectionStatus. instanceID: %s connectionStatus: %s", instanceID, model.ConnectionStatus_CONNECTED)
	}

	return nil
}

func (exec *VDCExecutor) startInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Push back to the state below in case of error.
	finState := model.InstanceState_STOPPED
	defer func() {
		err = model.Instances(ctx).UpdateState(instanceID, finState)
		if err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return err
	}

	err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STARTING)
	if err != nil {
		log.WithError(err).WithField("state", model.InstanceState_STOPPING).Error("Failed Instances.UpdateState")
		return err
	}

	log.Infof("Starting instance")
	err = hv.StartInstance()
	if err != nil {
		log.WithError(err).Error("Failed StartInstance")
		return err
	}
	log.Infof("Instance started successfully")
	// Here can bring the instance state to RUNNING finally.
	finState = model.InstanceState_RUNNING

	err = model.Instances(ctx).UpdateConnectionStatus(instanceID, model.ConnectionStatus_CONNECTED)

	if err != nil {
		return errors.Wrapf(err, "Couldn't update instance connectionStatus. instanceID: %s connectionStatus: %s", instanceID, model.ConnectionStatus_CONNECTED)
	}

	return nil
}

func (exec *VDCExecutor) stopInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Push back to the state below in case of error.
	finState := model.InstanceState_RUNNING
	defer func() {
		err = model.Instances(ctx).UpdateState(instanceID, finState)
		if err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return err
	}

	err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_STOPPING)
	if err != nil {
		log.WithError(err).WithField("state", model.InstanceState_STOPPING).Error("Failed Instances.UpdateState")
		return err
	}

	log.Infof("Stopping instance")
	err = hv.StopInstance()
	if err != nil {
		log.Error("Failed StopInstance")
		return err
	}
	log.Infof("Instance stopped successfully")
	// Here can bring the instance state to STOPPED finally.
	finState = model.InstanceState_STOPPED
	return nil
}

func (exec *VDCExecutor) rebootInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"hypervisor":  exec.hypervisorProvider.Name(),
	})

	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
		return err
	}

	// Push back to the state below in case of error.
	finState := model.InstanceState_RUNNING
	defer func() {
		err = model.Instances(ctx).UpdateState(instanceID, finState)
		if err != nil {
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return err
	}

	err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_REBOOTING)
	if err != nil {
		log.WithError(err).WithField("state", model.InstanceState_REBOOTING).Error("Failed Instances.UpdateState")
		return err
	}

	log.Infof("Rebooting instance")
	err = hv.RebootInstance()
	if err != nil {
		log.Error("Failed RebootInstance")
		return err
	}

	log.Infof("Instance rebooted successfully")
	finState = model.InstanceState_RUNNING

	return nil
}

func (exec *VDCExecutor) terminateInstance(driver exec.ExecutorDriver, instanceID string) error {
	log := log.WithFields(logrus.Fields{
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
		log.Errorln(err)
	}

	originalState := inst.GetLastState().GetState()

	// Push back to the state below in case of error.
	finState := model.InstanceState_RUNNING
	defer func() {
		err = model.Instances(ctx).UpdateState(instanceID, finState)
		if err != nil {
			log.WithError(err).WithField("state", finState).Error("Failed Instances.UpdateState")
		}
		model.Close(ctx)
	}()

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return err
	}

	err = model.Instances(ctx).UpdateState(instanceID, model.InstanceState_SHUTTINGDOWN)
	if err != nil {
		log.WithError(err).WithField("state", model.InstanceState_SHUTTINGDOWN).Error("Failed Instances.UpdateState")
		return err
	}

	// Trying to stop an already stopped container results in an error
	// causing the container to not get destroyed.

	if originalState != model.InstanceState_STOPPED {
		log.Infof("Shuttingdown instance")
		err = hv.StopInstance()
		if err != nil {
			log.Error("Failed StopInstance")
			return err
		}
	}

	err = hv.DestroyInstance()
	if err != nil {
		log.WithError(err).Error("Failed DestroyInstance")
		return err
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
	var err error

	switch command {
	case "start":
		err = exec.startInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to start instance")
		}
		err = exec.Failure(taskId.GetValue(), model.FailureMessage_FAILED_START)
		if err != nil {
			log.Errorln(err)
		}
	case "stop":
		err = exec.stopInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to stop instance")
			err = exec.Failure(taskId.GetValue(), model.FailureMessage_FAILED_STOP)
			if err != nil {
				log.Errorln(err)
			}
		}
	case "reboot":
		err = exec.rebootInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to reboot instance")
			err = exec.Failure(taskId.GetValue(), model.FailureMessage_FAILED_REBOOT)
			if err != nil {
				log.Errorln(err)
			}
		}
	case "destroy":
		var tstatus *mesos.TaskStatus
		err = exec.terminateInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to terminate instance")
			tstatus = &mesos.TaskStatus{
				TaskId: taskId,
				State:  mesos.TaskState_TASK_FAILED.Enum(),
			}
			err = exec.Failure(taskId.GetValue(), model.FailureMessage_FAILED_TERMINATE)
			if err != nil {
				log.Errorln(err)
			}
		} else {
			tstatus = &mesos.TaskStatus{
				TaskId: taskId,
				State:  mesos.TaskState_TASK_FINISHED.Enum(),
			}
		}
		if _, err := driver.SendStatusUpdate(tstatus); err != nil {
			log.WithError(err).Error("Couldn't send status update")
		}
	default:
		log.WithField("msg", msg).Errorln("FrameworkMessage unrecognized.")
	}
}

func (exec *VDCExecutor) Failure(instanceID string, failureMessage model.FailureMessage_ErrorType) error {
	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		return err
	}
	err = model.Instances(ctx).AddFailureMessage(instanceID, failureMessage)
	if err != nil {
		return err
	}
	model.Close(ctx)
	return nil
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
	viper.SetDefault("zookeeper.endpoint", "zk://localhost/openvdc")
	viper.SetDefault("executor-api.listen", "0.0.0.0:"+defaultExecutorAPIPort)
	viper.SetDefault("executor-api.advertise-ip", "")
	viper.SetDefault("console.ssh.listen", "0.0.0.0:"+defaultSSHPort)
	viper.SetDefault("console.ssh.advertise-ip", "")

	cobra.OnInitialize(initConfig)
	pfs := rootCmd.PersistentFlags()
	pfs.String("config", DefaultConfPath, "Load config file from the path")
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

	executorAPIListener, err := net.Listen("tcp", viper.GetString("executor-api.listen"))
	if err != nil {
		log.WithError(err).Fatalln("Faild to bind address for Executor gRPC API: ", viper.GetString("executor-api.listen"))
	}
	exposedExecutorAPIAddr := executorAPIListener.Addr().String()
	if viper.GetString("executor-api.advertise-ip") != "" {
		_, port, err := net.SplitHostPort(exposedExecutorAPIAddr)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse host:port: ", exposedExecutorAPIAddr)
		}
		exposedExecutorAPIAddr = net.JoinHostPort(viper.GetString("executor-api.advertise-ip"), port)
	}

	s := startExecutorAPIServer(ctx, executorAPIListener)
	defer s.GracefulStop()
	log.Infof("Listening Executor gRPC API on %s", executorAPIListener.Addr().String())
	log.Infof("Exposed Executor gRPC API on %s", exposedExecutorAPIAddr)

	sshListener, err := net.Listen("tcp", viper.GetString("console.ssh.listen"))
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen SSH on %s", sshListener.Addr().String())
	}
	defer sshListener.Close()
	exposedSSHAddr := sshListener.Addr().String()
	if viper.GetString("console.ssh.advertise-ip") != "" {
		_, port, err := net.SplitHostPort(exposedSSHAddr)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse host:port: ", exposedSSHAddr)
		}
		exposedSSHAddr = net.JoinHostPort(viper.GetString("console.ssh.advertise-ip"), port)
	}

	clusterNode := &model.ExecutorNode{
		GrpcAddr: exposedExecutorAPIAddr,
		Console: &model.Console{
			Type:     model.Console_SSH,
			BindAddr: exposedSSHAddr,
		},
	}
	vdcExecutor := newVDCExecutor(ctx, clusterNode)

	sshd := NewSSHServer(vdcExecutor)
	if err := sshd.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to setup SSH Server")
	}
	go sshd.Run(sshListener)
	log.Infof("Listening SSH on %s", sshListener.Addr().String())
	log.Infof("Exposed SSH on %s", exposedSSHAddr)

	dconfig := exec.DriverConfig{
		Executor: vdcExecutor,
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
	logrus.SetFormatter(&cmd.LogFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	rootCmd.AddCommand(cmd.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
