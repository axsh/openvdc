package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/axsh/openvdc"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	mesosutil "github.com/mesos/mesos-go/mesosutil"
	"golang.org/x/net/context"
)

var log = logrus.WithField("context", "vdc-executor")

type VDCExecutor struct {
	hypervisorProvider hypervisor.HypervisorProvider
}

func newVDCExecutor(provider hypervisor.HypervisorProvider) *VDCExecutor {
	return &VDCExecutor{
		hypervisorProvider: provider,
	}
}

func (exec *VDCExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	log.Infoln("Registered Executor on slave ", slaveInfo.GetHostname())
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
		log.Errorln("Couldn't send status update", err)
	}

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
	case "reboot":
		err = exec.rebootInstance(driver, taskId.GetValue())
		if err != nil {
			log.WithError(err).Error("Failed to reboot instance")
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

func init() {
	viper.SetDefault("hypervisor.driver", "null")
	viper.SetDefault("zookeeper.endpoint", "zk://localhost/openvdc")

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

	dconfig := exec.DriverConfig{
		Executor: newVDCExecutor(provider),
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
	rootCmd.AddCommand(openvdc.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
