package main

import (
	"flag"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"

	"github.com/axsh/openvdc/api/executor"
	"github.com/axsh/openvdc/hypervisor"
	"github.com/axsh/openvdc/model"
	mesosutil "github.com/mesos/mesos-go/mesosutil"
	"golang.org/x/net/context"
)

// Build time constant variables from -ldflags
var (
	version   string
	sha       string
	builddate string
	goversion string
)

var log = logrus.WithField("context", "vdc-executor")

type VDCExecutor struct {
	hypervisorProvider hypervisor.HypervisorProvider
	ctx                context.Context
	executorAPI        *executor.ExecutorAPIServer
}

func newVDCExecutor(ctx context.Context, provider hypervisor.HypervisorProvider, svr *executor.ExecutorAPIServer) *VDCExecutor {
	return &VDCExecutor{
		hypervisorProvider: provider,
		ctx:                ctx,
		executorAPI:        svr,
	}
}

func (exec *VDCExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	log.Infoln("Registered Executor on slave ", slaveInfo.GetHostname())

	node := &model.ClusterNode{
		Id:       slaveInfo.GetId().GetValue(),
		GrpcAddr: exec.executorAPI.Listener().Addr().String(),
		Console: &model.Console{
			Type:     model.Console_SSH,
			BindAddr: "",
		},
	}
	err := model.Cluster(exec.ctx).Register(node)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infoln("Registered on OpenVDC cluster service: ", node)
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

	ctx, err := model.Connect(context.Background(), []string{*zkAddr})
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

	ctx, err := model.Connect(context.Background(), []string{*zkAddr})
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

	ctx, err := model.Connect(context.Background(), []string{*zkAddr})
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

	ctx, err := model.Connect(context.Background(), []string{*zkAddr})
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

var (
	hypervisorName = flag.String("hypervisor", "null", "")
	zkAddr         = flag.String("zk", "127.0.0.1:2181", "Zookeeper address")
)

func startExecutorAPIServer(ctx context.Context) *executor.ExecutorAPIServer {
	laddr := "0.0.0.0:19372"
	lis, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalln("Faild to bind address for gRPC API: ", laddr)
	}
	log.Println("Listening gRPC API on: ", laddr)
	s := executor.NewExecutorAPIServer(*zkAddr, ctx)
	go s.Serve(lis)
	return s
}

func init() {
	flag.Parse()
}

func main() {
	provider, ok := hypervisor.FindProvider(*hypervisorName)
	if ok == false {
		log.Fatalln("Unknown hypervisor name:", hypervisorName)
	}
	log.Infof("Initializing executor: hypervisor %s\n", provider.Name())

	ctx, err := model.ClusterConnect(context.Background(), []string{*zkAddr})
	if err != nil {
		log.Fatal(err)
		return
	}
	defer func() {
		err := model.ClusterClose(ctx)
		if err != nil {
			log.Error(err)
		}
	}()
	s := startExecutorAPIServer(ctx)
	defer s.GracefulStop()

	dconfig := exec.DriverConfig{
		Executor: newVDCExecutor(ctx, provider, s),
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
