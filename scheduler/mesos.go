package scheduler

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/api"
	"github.com/gogo/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/auth/sasl"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
)

const (
	CPUS_PER_EXECUTOR = 0.01
	CPUS_PER_TASK     = 1
	MEM_PER_EXECUTOR  = 64
	MEM_PER_TASK      = 64
)

const ExecutorPath = "openvdc-executor"

var (
	taskCount = 10
)

type VDCScheduler struct {
	executor      *mesos.ExecutorInfo
	tasksLaunched int
	tasksFinished int
	tasksErrored  int
	totalTasks    int
	offerChan     api.APIOffer
	listenAddr    string
}

func newVDCScheduler(ch api.APIOffer, listenAddr string) *VDCScheduler {
	exec := &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID("vdc-hypervisor-null"),
		Name:       proto.String("VDC Executor"),
		Source:     proto.String("go_test"),
		Command: &mesos.CommandInfo{
			Value: proto.String(fmt.Sprintf("%s -logtostderr=true -hypervisor=null", ExecutorPath)),
		},
		Resources: []*mesos.Resource{
			util.NewScalarResource("cpus", CPUS_PER_EXECUTOR),
			util.NewScalarResource("mem", MEM_PER_EXECUTOR),
		},
	}

	return &VDCScheduler{
		executor:   exec,
		totalTasks: taskCount,
		offerChan:  ch,
		listenAddr: listenAddr,
	}
}

func (sched *VDCScheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Println("Framework Registered with Master ", masterInfo)
}

func (sched *VDCScheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Println("Framework Re-Registered with Master ", masterInfo)
}

func (sched *VDCScheduler) Disconnected(sched.SchedulerDriver) {
	log.Println("disconnected from master")
}

func (sched *VDCScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log := log.WithFields(log.Fields{"offers": len(offers)})
	select {
	case s := <-sched.offerChan:
		log.Infoln("Process this offer as new request arrived.")
		imageName := s.ImageName
		hostName := s.HostName
		taskType := s.TaskType

		fmt.Println("Scheduler, ImageName: ", imageName)
		fmt.Println("Scheduler, HostName: ", hostName)
		fmt.Println("Scheduler, TaskType: ", taskType)

		var clientCommands string
		if imageName != "" {
			clientCommands = "imageName=" + imageName
		}

		if hostName != "" {
			clientCommands = clientCommands + "&hostName=" + hostName
		}

		if taskType != "" {
			clientCommands = clientCommands + "&taskType=" + taskType
		}

		sched.processOffers(driver, offers, clientCommands, hostName)
	default:
		log.Debugln("Skip offer since no allocation requests.")
		for _, offer := range offers {
			stat, err := driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
			if err != nil {
				log.Errorln(err)
			}
			log = log.WithField("mesos.Status", stat.String())
			// Assert returned status.
			if stat != mesos.Status_DRIVER_RUNNING {
				log.Fatalln("Invalid status")
			}
		}
	}
}

func (sched *VDCScheduler) processOffers(driver sched.SchedulerDriver, offers []*mesos.Offer, clientCommands string, hostName string) {
	if (sched.tasksLaunched - sched.tasksErrored) >= sched.totalTasks {
		log.Println("All tasks are already launched: decline offer")
		for _, offer := range offers {
			driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(120)})
		}
		return
	}

	for _, offer := range offers {
		cpuResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "cpus"
		})
		cpus := 0.0
		for _, res := range cpuResources {
			cpus += res.GetScalar().GetValue()
		}

		memResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "mem"
		})
		mems := 0.0
		for _, res := range memResources {
			mems += res.GetScalar().GetValue()
		}

		log.Println("Received Offer <", offer.Id.GetValue(), "> with cpus=", cpus, " mem=", mems)

		remainingCpus := cpus
		remainingMems := mems

		if len(offer.ExecutorIds) == 0 {
			remainingCpus -= CPUS_PER_EXECUTOR
			remainingMems -= MEM_PER_EXECUTOR
		}

		var tasks []*mesos.TaskInfo

		sched.tasksLaunched++
		taskId := util.NewTaskID(hostName)

		task := &mesos.TaskInfo{
			Name:     proto.String("VDC" + "_" + taskId.GetValue()),
			TaskId:   taskId,
			SlaveId:  offer.SlaveId,
			Data:     []byte(clientCommands),
			Executor: sched.executor,
			Resources: []*mesos.Resource{
				util.NewScalarResource("cpus", CPUS_PER_TASK),
				util.NewScalarResource("mem", MEM_PER_TASK),
			},
		}
		log.Printf("Prepared task: %s with offer %s for launch\n", task.GetName(), offer.Id.GetValue())

		tasks = append(tasks, task)
		remainingCpus -= CPUS_PER_TASK
		remainingMems -= MEM_PER_TASK

		log.Println("Launching ", len(tasks), "tasks for offer", offer.Id.GetValue())
		// TODO: Replace with AcceptOffers(). https://issues.apache.org/jira/browse/MESOS-2955
		driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	}
}

func (sched *VDCScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Println("Framework Resource Offers from master", status)

	if status.GetState() == mesos.TaskState_TASK_FINISHED {
		sched.tasksFinished++
		driver.ReviveOffers()
	}

	if status.GetState() == mesos.TaskState_TASK_LOST ||
		status.GetState() == mesos.TaskState_TASK_ERROR ||
		status.GetState() == mesos.TaskState_TASK_FAILED ||
		status.GetState() == mesos.TaskState_TASK_KILLED {
		sched.tasksErrored++
	}
}

func (sched *VDCScheduler) OfferRescinded(_ sched.SchedulerDriver, oid *mesos.OfferID) {
	log.Fatalf("offer rescinded: %v", oid)
}

func (sched *VDCScheduler) FrameworkMessage(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, msg string) {
	log.Fatalf("framework message from executor %q slave %q: %q", eid, sid, msg)
}

func (sched *VDCScheduler) SlaveLost(_ sched.SchedulerDriver, sid *mesos.SlaveID) {
	log.Fatalf("slave lost: %v", sid)
}

func (sched *VDCScheduler) ExecutorLost(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, code int) {
	log.Fatalf("executor %q lost on slave %q code %d", eid, sid, code)
}

func (sched *VDCScheduler) Error(_ sched.SchedulerDriver, err string) {
	log.Fatalf("Scheduler received error: %v", err)
}

func startAPIServer(laddr string, ch api.APIOffer, zkAddr string, driver sched.SchedulerDriver) *api.APIServer {
	lis, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalln("Faild to bind address for gRPC API: ", laddr)
	}
	log.Println("Listening gRPC API on: ", laddr)
	s := api.NewAPIServer(ch, zkAddr, driver)
	go s.Serve(lis)
	return s
}

func Run(listenAddr string, apiListenAddr string, mesosMasterAddr string, zkAddr string) {
	ch := make(api.APIOffer)

	fwinfo := &mesos.FrameworkInfo{
		User: proto.String(""), // Mesos-go will fill in user.
		Name: proto.String("VDC Scheduler"),
	}

	cred := &mesos.Credential{
		Principal: proto.String(""),
		Secret:    proto.String(""),
	}

	cred = nil
	bindingAddrs, err := net.LookupIP(listenAddr)
	if err != nil {
		log.Fatalln("Invalid Address to -listen option: ", err)
	}
	config := sched.DriverConfig{
		Scheduler:      newVDCScheduler(ch, listenAddr),
		Framework:      fwinfo,
		Master:         mesosMasterAddr,
		Credential:     cred,
		BindingAddress: bindingAddrs[0],
		WithAuthContext: func(ctx context.Context) context.Context {
			ctx = auth.WithLoginProvider(ctx, sasl.ProviderName)
			ctx = sasl.WithBindingAddress(ctx, bindingAddrs[0])
			return ctx
		},
	}
	driver, err := sched.NewMesosSchedulerDriver(config)
	if err != nil {
		log.Fatalln("Unable to create a SchedulerDriver ", err.Error())
	}

	apiServer := startAPIServer(apiListenAddr, ch, zkAddr, driver)
	defer func() {
		apiServer.GracefulStop()
	}()

	if stat, err := driver.Run(); err != nil {
		log.Printf("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
	}
}
