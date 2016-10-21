package main

import (
	"flag"
	"fmt"
	"log"
	"net"

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
	numOfTasks = 1
)

var (
	bindingIPv4        = flag.String("listen", "localhost", "Bind address")
	mesosMasterAddress = flag.String("master", "localhost:5050", "Mesos Master node")
	taskCount            =  flag.String("task-count", "numOfTasks", "Number of tasks to run")
)

type VDCScheduler struct {
	executor *mesos.ExecutorInfo
	tasksLaunched int
        tasksFinished int
        tasksErrored  int
        totalTasks    int
}

func newVDCScheduler() *VDCScheduler {
	executorUris := []*mesos.CommandInfo_URI{}

        uri, executorCmd := serveExecutorArtifact(*executorPath)
        executorUris = append(executorUris, &mesos.CommandInfo_URI{Value: uri, Executable: proto.Bool(true)})

        //Pass flags to executor
        v := 0
        if f := flag.Lookup("v"); f != nil && f.Value != nil {
                if vstr := f.Value.String(); vstr != "" {
                        if vi, err := strconv.ParseInt(vstr, 10, 32); err == nil {
                                v = int(vi)
                        }
                }
        }

	executorCommand := fmt.Sprintf("./%s -logtostderr=true -v=%d -slow_tasks=false", executorCmd, v)
        go http.ListenAndServe(fmt.Sprintf("%s:%d", *bindingIPv4, *artifactPort), nil)

	exec := &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID("default"),
		Name:       proto.String("Test Executor (Go)"),
		Source:     proto.String("go_test"),
		Command: &mesos.CommandInfo{
			Value: proto.String(executorCommand),
			Uris:  executorUris,
		},
		Resources: []*mesos.Resource{
			util.NewScalarResource("cpus", CPUS_PER_EXECUTOR),
			util.NewScalarResource("mem", MEM_PER_EXECUTOR),
		},
	}

	total, err := strconv.Atoi(*taskCount)

        if err != nil {
                log.Println("ERROR: Taskcount not specified. Setting it to 4.")
                total = 4
        }

        return &VDCScheduler{executor: exec, totalTasks: total}
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
	if (sched.tasksLaunched - sched.tasksErrored) >= sched.totalTasks {
                log.Println("All tasks are already launched: decline offer")
                ids := make([]*mesos.OfferID, len(offers))
                for i, offer := range offers {
                        ids[i] = offer.Id
                }
                driver.LaunchTasks(ids, []*mesos.TaskInfo{}, &mesos.Filters{RefuseSeconds: proto.Float64(120)})
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

                        task := &mesos.TaskInfo{
                                Name:     proto.String("go-task-" + taskId.GetValue()),
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
                driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	}
}

func (sched *VDCScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Println("Framework Resource Offers from master", status)

        if status.GetState() == mesos.TaskState_TASK_FINISHED {
                sched.tasksFinished++
                driver.ReviveOffers()
        }

        if sched.tasksFinished >= sched.totalTasks {
                log.Println("Tasks completed, stopping framework.")
                driver.Stop(false)
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

func init() {
	flag.Parse()
}

func serveExecutorArtifact(path string) (*string, string) {
        serveFile := func(pattern string, filename string) {
                http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
                        http.ServeFile(w, r, filename)
                })
        }

        pathSplit := strings.Split(path, "/")
        var base string
        if len(pathSplit) > 0 {
                base = pathSplit[len(pathSplit)-1]
        } else {
                base = path
        }
        serveFile("/"+base, path)

        hostURI := fmt.Sprintf("http://%s:%d/%s", *bindingIPv4, *artifactPort, base)
        log.Println("Hosting artifact '%s' at '%s'", path, hostURI)

        return &hostURI, base
}

func main() {
	fwinfo := &mesos.FrameworkInfo{
		User: proto.String(""), // Mesos-go will fill in user.
		Name: proto.String("VDC Scheduler"),
	}

	cred := &mesos.Credential{
		Principal: proto.String(""),
		Secret:    proto.String(""),
	}

	cred = nil
	bindingAddrs, err := net.LookupIP(*bindingIPv4)
	if err != nil {
		log.Fatalln("Invalid Address to -listen option: ", err)
	}
	config := sched.DriverConfig{
		Scheduler:      newVDCScheduler(),
		Framework:      fwinfo,
		Master:         *mesosMasterAddress,
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
