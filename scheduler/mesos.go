package scheduler

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"

	"net/http"
	"strconv"
	"strings"

	"github.com/axsh/openvdc/api"
	"github.com/gogo/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/auth/sasl"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
)

var running bool = false
var executorCmd string
var executorUris []*mesos.CommandInfo_URI
var executorCount int = 1

const (
	CPUS_PER_EXECUTOR   = 0.01
	CPUS_PER_TASK       = 1
	MEM_PER_EXECUTOR    = 64
	MEM_PER_TASK        = 64
	defaultArtifactPort = 12345
)

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
	artifactPort  uint16
	executorPath  string
}

func newVDCScheduler(ch api.APIOffer, listenAddr string) *VDCScheduler {

	go http.ListenAndServe(fmt.Sprintf("%s:%d", listenAddr, defaultArtifactPort), nil)

	exec := &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID("default"),
		Name:       proto.String("VDC Executor"),
		Source:     proto.String("go_test"),
		Command:    &mesos.CommandInfo{},
		Resources: []*mesos.Resource{
			util.NewScalarResource("cpus", CPUS_PER_EXECUTOR),
			util.NewScalarResource("mem", MEM_PER_EXECUTOR),
		},
	}

	return &VDCScheduler{
		executor:     exec,
		totalTasks:   taskCount,
		offerChan:    ch,
		listenAddr:   listenAddr,
		artifactPort: defaultArtifactPort,
		executorPath: "executor",
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
	select {
	case s := <-sched.offerChan:

		imageName := s.ImageName
		hostName := s.HostName

		fmt.Println("Scheduler, ImageName: ", imageName)
		fmt.Println("Scheduler, HostName: ", hostName)

		var clientCommands string
		if imageName != "" {
			clientCommands = " -imageName=" + imageName
		}

		if hostName != "" {
			clientCommands = clientCommands + " -hostName=" + hostName
		}

		sched.processOffers(driver, offers, clientCommands)
	default:
		log.Println("Skip offer since no allocation requests.", offers)
		for _, offer := range offers {
			stat, err := driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
			if err != nil {
				log.Println(err)
			}
			log.Println(stat)
		}
	}
}

func (sched *VDCScheduler) processOffers(driver sched.SchedulerDriver, offers []*mesos.Offer, clientCommands string) {
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

		for (sched.tasksLaunched-sched.tasksErrored) < sched.totalTasks &&
			CPUS_PER_TASK <= remainingCpus &&
			MEM_PER_TASK <= remainingMems {

			sched.tasksLaunched++

			taskId := &mesos.TaskID{
				Value: proto.String(strconv.Itoa(sched.tasksLaunched)),
			}

			var uri *string

			//Don't re-delcare these if the scheduler is arleady up and running.
			if running == false {
				uri, executorCmd = sched.serveExecutorArtifact()
				executorUris = append(executorUris, &mesos.CommandInfo_URI{Value: uri, Executable: proto.Bool(true)})
				running = true
			}

			sched.executor.ExecutorId = util.NewExecutorID(strconv.Itoa(executorCount))
			executorCount += 1
			executorCommand := fmt.Sprintf("./%s -logtostderr=true -slow_tasks=false"+clientCommands, executorCmd)

			sched.executor.Command = &mesos.CommandInfo{
				Value: proto.String(executorCommand),
				Uris:  executorUris,
			}

			fmt.Println("----------------------------------")
			fmt.Printf("%+v", sched.executor.Command)
			fmt.Println("----------------------------------")

			task := &mesos.TaskInfo{
				Name:     proto.String("VDC" + "_" + taskId.GetValue()),
				TaskId:   taskId,
				SlaveId:  offer.SlaveId,
				Executor: sched.executor,
				Resources: []*mesos.Resource{
					util.NewScalarResource("cpus", CPUS_PER_TASK),
					util.NewScalarResource("mem", MEM_PER_TASK),
				},
			}
			log.Println("Prepared task: %s with offer %s for launch\n", task.GetName(), offer.Id.GetValue())

			tasks = append(tasks, task)
			remainingCpus -= CPUS_PER_TASK
			remainingMems -= MEM_PER_TASK
		}

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

func (sched *VDCScheduler) serveExecutorArtifact() (*string, string) {
	serveFile := func(pattern string, filename string) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filename)
		})
	}

	pathSplit := strings.Split(sched.executorPath, "/")
	var base string
	if len(pathSplit) > 0 {
		base = pathSplit[len(pathSplit)-1]
	} else {
		base = sched.executorPath
	}
	serveFile("/"+base, sched.executorPath)

	hostURI := fmt.Sprintf("http://%s:%d/%s", sched.listenAddr, sched.artifactPort, base)
	log.Println("Hosting artifact '%s' at '%s'", sched.executorPath, hostURI)

	return &hostURI, base
}

func startAPIServer(laddr string, ch api.APIOffer) *api.APIServer {
	lis, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalln("Faild to bind address for gRPC API: ", laddr)
	}
	log.Println("Listening gRPC API on: ", laddr)
	s := api.NewAPIServer(ch)
	go s.Serve(lis)
	return s
}

func Run(listenAddr string, apiListenAddr string, mesosMasterAddr string) {
	ch := make(api.APIOffer)
	apiServer := startAPIServer(apiListenAddr, ch)
	defer func() {
		apiServer.GracefulStop()
	}()
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
	if stat, err := driver.Run(); err != nil {
		log.Printf("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
	}
}
