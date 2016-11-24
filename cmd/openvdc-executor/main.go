package main

import (
	"net/url"
	"strings"
	"time"

	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"

	"flag"

	"github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/util"
	mesosutil "github.com/mesos/mesos-go/mesosutil"
)

// Build time constant variables from -ldflags
var (
	version   string
	sha       string
	builddate string
	goversion string
)

var theTaskInfo mesos.TaskInfo

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

func newTask(theHostName string, taskType string, exec *VDCExecutor) {

	hp := exec.hypervisorProvider
	hp.SetName(theHostName)
	hvd, err := hp.CreateDriver()

	if err != nil {
		log.Errorln("Hypervisor driver error", err)
		return
	}

	switch taskType {
	case "create":
		err = hvd.CreateInstance()
		if err != nil {
			log.Errorln("Error creating instance")
		}
	case "destroy":
		err = hvd.DestroyInstance()
		if err != nil {
			log.Errorln("Error destroying instance")
		}
	case "run":
		err = hvd.StartInstance()
		if err != nil {
			log.Errorln("Error running instance")
		}
	case "stop":
		err = hvd.StopInstance()
		if err != nil {
			log.Errorln("Error stopping instance")
		}
	case "console":
		err = hvd.InstanceConsole()
		if err != nil {
			log.Errorln("Error connecting to instance")
		}
	default:
		log.Errorln("Invalid task name")
	}
}

func (exec *VDCExecutor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	log.Infoln("Launching task", taskInfo.GetName(), "with command", taskInfo.Command.GetValue())

	theTaskInfo = *taskInfo

	runStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	}
	_, err := driver.SendStatusUpdate(runStatus)
	if err != nil {
		log.Errorln("Couldn't send status update", err)
	}


	b := taskInfo.GetData()
	s := string(b[:])

	values, err := url.ParseQuery(s)

	if err != nil {
		panic(err)
	}

	imageName := values.Get("imageName")
	hostName := values.Get("hostName")
	taskType := values.Get("taskType")

	log.Infoln("ImageName: "+imageName+", HostName: "+hostName, "TaskType: "+taskType)

	newTask(hostName, taskType, exec)

	/*hv, err := exec.hypervisorProvider.CreateDriver()
	if err != nil {
		finState = mesos.TaskState_TASK_FAILED
		return
	}

	err = hv.CreateInstance()
	if err != nil {
		finState = mesos.TaskState_TASK_FAILED
		return
	}
	err = hv.StartInstance()
	if err != nil {
		finState = mesos.TaskState_TASK_FAILED
		return
	}*/
}

func DestroyTask(driver exec.ExecutorDriver, taskId *mesos.TaskID) {

	finState := mesos.TaskState_TASK_FINISHED

	log.Infoln("Finishing task", theTaskInfo.GetName())
	finStatus := &mesos.TaskStatus{
		TaskId: taskId,
		State:  finState.Enum(),
	}

	if _, err := driver.SendStatusUpdate(finStatus); err != nil {
		log.Infoln("ERROR: Couldn't send status update", err)
	}
	log.Infoln("Task finished", theTaskInfo.GetName())
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
	case "destroy":
		DestroyTask(driver, taskId)
	default:
		log.Errorln("FrameworkMessage unrecognized.")
	}
}

func (exec *VDCExecutor) Shutdown(driver exec.ExecutorDriver) {
	log.Infoln("Shutting down the executor")
}

func (exec *VDCExecutor) Error(driver exec.ExecutorDriver, err string) {
	log.Errorln("Got error message:", err)
}

var hypervisorName = flag.String("hypervisor", "null", "")

func init() {
	flag.Parse()
}

func main() {
	util.SetupLog()

	provider, ok := hypervisor.FindProvider(*hypervisorName)
	if ok == false {
		log.Fatalln("Unknown hypervisor name:", hypervisorName)
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
