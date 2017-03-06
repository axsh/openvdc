package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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

	runStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	}
	_, err := driver.SendStatusUpdate(runStatus)
	if err != nil {
		log.WithError(err).Error("Couldn't send status update")
	}

	instanceState, containerState, err := exec.getStates(driver, taskInfo.GetTaskId().GetValue())

	if err != nil {
		log.WithError(err).Error("Failed to get instance states")
	}

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
		}
	} else {
		err = exec.recoverInstance(driver, taskInfo.GetTaskId().GetValue(), instanceState, containerState)
	}
}

func (exec *VDCExecutor) recoverInstance(driver exec.ExecutorDriver, instanceID string, instanceState model.InstanceState, containerState hypervisor.ContainerState) error {
	switch instanceState.State {
	case model.InstanceState_STARTING:

	case model.InstanceState_RUNNING:
		if containerState == hypervisor.ContainerState_STOPPED {
			hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
			if err != nil {
				return errors.Wrapf(err, "Hypervisorprovider failed to create driver. InstanceID:  %s", instanceID)
			}

			err = hv.StartInstance()
			if err != nil {
				return errors.Wrapf(err, "Failed to start instance:  %s", instanceID)
			}
		}
	case model.InstanceState_STOPPING:

	case model.InstanceState_STOPPED:

	case model.InstanceState_SHUTTINGDOWN:

	case model.InstanceState_TERMINATED:
	}
	return nil
}

func (exec *VDCExecutor) getStates(driver exec.ExecutorDriver, instanceID string) (model.InstanceState, hypervisor.ContainerState, error) {
	ctx, err := model.Connect(context.Background(), &zkAddr)
	if err != nil {
		return model.InstanceState{}, hypervisor.ContainerState_NONE, errors.Wrapf(err, "Failed model.Connect")
	}

	instance, err := model.Instances(ctx).FindByID(instanceID)

	if err != nil {
		return model.InstanceState{}, hypervisor.ContainerState_NONE, errors.Wrapf(err, "Failed to fetch instance %s", instanceID)
	}

	instanceState := instance.GetLastState()

	hv, err := exec.hypervisorProvider.CreateDriver(instanceID)
	if err != nil {
		return model.InstanceState{}, hypervisor.ContainerState_NONE, errors.Wrapf(err, "Failed to create hypervisor driver. InstanceID: %s", instanceID)
	}

	containerState, err := hv.GetContainerState(instance)
	if err != nil {
		return model.InstanceState{}, hypervisor.ContainerState_NONE, errors.Wrapf(err, "Hypervisor failed to get container state. InstanceID: %s", instanceID)
	}

	return *instanceState, containerState, nil
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
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
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
	res, err := inst.Resource(ctx)
	if err != nil {
		log.WithError(err).Error("Failed Instances.Resource")
		return err
	}

	log.Infof("Creating instance")
	err = hv.CreateInstance(inst, res.ResourceTemplate())
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
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
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
		log.Error("Failed StartInstance")
		return err
	}
	log.Infof("Instance started successfully")
	// Here can bring the instance state to RUNNING finally.
	finState = model.InstanceState_RUNNING
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
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
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
			log.WithField("state", finState).Error("Failed Instances.UpdateState")
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
		log.Error("Failed DestroyInstance")
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
	case "stop":
		err = exec.stopInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to stop instance")
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

var defaultSSHPortRange = [2]int{29876, 39876}

func startExecutorAPIServer(ctx context.Context, listener net.Listener) *executor.ExecutorAPIServer {
	s := executor.NewExecutorAPIServer(&zkAddr, ctx)
	go s.Serve(listener)
	return s
}

func init() {
	viper.SetDefault("hypervisor.driver", "null")
	viper.SetDefault("zookeeper.endpoint", "zk://localhost/openvdc")
	viper.SetDefault("executor-api.listen", "0.0.0.0:19372")
	viper.SetDefault("executor-api.advertise-ip", "")
	viper.SetDefault("console.ssh.listen", "")
	viper.SetDefault("console.ssh.advertise-ip", "")

	cobra.OnInitialize(initConfig)
	pfs := rootCmd.PersistentFlags()
	pfs.String("config", DefaultConfPath, "Load config file from the path")
	pfs.String("hypervisor", viper.GetString("hypervisor.driver"), "Hypervisor driver name")
	viper.BindPFlag("hypervisor.driver", pfs.Lookup("hypervisor"))
	pfs.String("zk", viper.GetString("zookeeper.endpoint"), "Zookeeper address")
	viper.BindPFlag("zookeeper.endpoint", pfs.Lookup("zk"))
}

func randomPort(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
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
		log.Fatalf("Failed to load config %s: %v", viper.ConfigFileUsed(), err)
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
	log.Infof("Initializing executor: hypervisor %s\n", provider.Name())

	ctx, err := model.ClusterConnect(context.Background(), &zkAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connect to cluster service %s", zkAddr.String())
	}
	defer func() {
		err := model.ClusterClose(ctx)
		if err != nil {
			log.Error(err)
		}
	}()

	executorAPIListener, err := net.Listen("tcp", viper.GetString("executor-api.listen"))
	if err != nil {
		log.Fatalln("Faild to bind address for Executor gRPC API: ", viper.GetString("executor-api.listen"))
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

	sshPort := strconv.Itoa(randomPort(defaultSSHPortRange[0], defaultSSHPortRange[1]))
	sshListenIP := "0.0.0.0"
	if viper.GetString("console.ssh.listen") != "" {
		var port string
		sshListenIP, port, err = net.SplitHostPort(viper.GetString("console.ssh.listen"))
		if err != nil {
			log.WithError(err).Fatal("Failed to parse host:port: ", viper.GetString("console.ssh.listen"))
		}
		sshPort = port
	}
	sshListenAddr := net.JoinHostPort(sshListenIP, sshPort)
	sshListener, err := net.Listen("tcp", sshListenAddr)
	if err != nil {
		log.WithError(err).Fatalf("Failed to listen SSH on %s", sshListenAddr)
	}
	defer sshListener.Close()

	sshd := NewSSHServer(provider)
	if err := sshd.Setup(); err != nil {
		log.WithError(err).Fatal("Failed to setup SSH Server")
	}
	go sshd.Run(sshListener)
	log.Infof("Listening SSH on %s", sshListenAddr)
	exposedSSHAddr := sshListener.Addr().String()
	if viper.GetString("console.ssh.advertise-ip") != "" {
		exposedSSHAddr = net.JoinHostPort(viper.GetString("console.ssh.advertise-ip"), sshPort)
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
		log.Fatalln("Couldn't create ExecutorDriver ", err)
	}

	_, err = driver.Start()
	if err != nil {
		log.Fatalln("ExecutorDriver wasn't able to start: ", err)
	}
	log.Infoln("Process running")

	_, err = driver.Join()
	if err != nil {
		log.Fatalln("Something went wrong with the driver: ", err)
	}
	log.Infoln("Executor shutting down")
}

func main() {
	logrus.SetFormatter(&cmd.LogFormatter{})
	rootCmd.AddCommand(cmd.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
