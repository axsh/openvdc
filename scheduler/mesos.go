package scheduler

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/axsh/openvdc/api"
	"github.com/axsh/openvdc/model"
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

var FrameworkInfo = &mesos.FrameworkInfo{
	User: proto.String(""),
	Name: proto.String("OpenVDC"),
}

const ExecutorPath = "openvdc-executor"

var (
	taskCount = 10
)

type VDCScheduler struct {
	tasksLaunched int
	tasksFinished int
	tasksErrored  int
	totalTasks    int
	listenAddr    string
	zkAddr        string
	ctx           context.Context
}

func newVDCScheduler(ctx context.Context, listenAddr string, zkAddr string) *VDCScheduler {
	return &VDCScheduler{
		totalTasks: taskCount,
		listenAddr: listenAddr,
		zkAddr:     zkAddr,
		ctx:        ctx,
	}
}

func (sched *VDCScheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Println("Framework Registered with Master ", masterInfo)
	node := &model.ClusterNode{
		Id: "scheduler",
	}
	err := model.Cluster(sched.ctx).Register(node)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infoln("Registered on OpenVDC cluster service: ", node)
}

func (sched *VDCScheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Println("Framework Re-Registered with Master ", masterInfo)
}

func (sched *VDCScheduler) Disconnected(sched.SchedulerDriver) {
	log.Println("disconnected from master")
}

func (sched *VDCScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log := log.WithFields(log.Fields{"offers": len(offers)})

	ctx, err := model.Connect(context.Background(), []string{sched.zkAddr})
	if err != nil {
		log.WithError(err).Error("Failed to connecto to datasource: ", sched.zkAddr)
	} else {
		defer model.Close(ctx)
		err = sched.processOffers(driver, offers, ctx)
		if err != nil {
			log.WithError(err).Error("Failed to process offers")
		}
	}
}

func (sched *VDCScheduler) processOffers(driver sched.SchedulerDriver, offers []*mesos.Offer, ctx context.Context) error {
	queued, err := model.Instances(ctx).FilterByState(model.InstanceState_QUEUED)
	if err != nil {
		return err
	}

	if len(queued) == 0 {
		log.Infoln("Skip offers since no allocation requests.")
		for _, offer := range offers {
			_, err := driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
			if err != nil {
				log.WithError(err).Error("Failed to response DeclineOffer.")
			}
		}
		return nil
	}

	getHypervisorName := func(offer *mesos.Offer) string {
		for _, attr := range offer.Attributes {
			if attr.GetName() == "hypervisor" &&
				attr.GetType() == mesos.Value_TEXT {
				return attr.GetText().GetValue()
			}
		}
		return ""
	}

	findMatching := func(i *model.Instance) *mesos.Offer {
		for _, offer := range offers {
			hypervisorName := getHypervisorName(offer)
			if hypervisorName == "" {
				continue
			}
			r, err := i.Resource(ctx)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"instance_id": i.GetId(),
					"resource_id": i.GetResourceId(),
				}).Error("Failed to retrieve resource object")
				continue
			}
			// TODO: Avoid type switch to find template types.
			switch t := r.GetTemplate().GetItem(); t.(type) {
			case *model.Template_Lxc:
				if hypervisorName == "lxc" {
					return offer
				}
			case *model.Template_Null:
				if hypervisorName == "null" {
					return offer
				}
			default:
				log.Warnf("Unknown template type")
			}
		}
		return nil
	}

	tasks := []*mesos.TaskInfo{}
	acceptIDs := []*mesos.OfferID{}
	for _, i := range queued {
		found := findMatching(i)
		if found == nil {
			continue
		}
		hypervisorName := getHypervisorName(found)
		log.WithFields(log.Fields{
			"instance_id": i.GetId(),
			"hypervisor":  hypervisorName,
		}).Info("Found matching offer")

		executor := &mesos.ExecutorInfo{
			ExecutorId: util.NewExecutorID(fmt.Sprintf("vdc-hypervisor-%s", hypervisorName)),
			Name:       proto.String("VDC Executor"),
			Command: &mesos.CommandInfo{
				Value: proto.String(fmt.Sprintf("%s -logtostderr=true -hypervisor=%s -zk=%s",
					ExecutorPath, hypervisorName, sched.zkAddr)),
			},
		}

		taskId := util.NewTaskID(i.GetId())
		task := &mesos.TaskInfo{
			Name:     proto.String("VDC" + "_" + taskId.GetValue()),
			TaskId:   taskId,
			SlaveId:  found.SlaveId,
			Data:     []byte("instance_id=" + i.GetId()),
			Executor: executor,
			Resources: []*mesos.Resource{
				util.NewScalarResource("cpus", CPUS_PER_TASK),
				util.NewScalarResource("mem", MEM_PER_TASK),
			},
		}

		tasks = append(tasks, task)
		acceptIDs = append(acceptIDs, found.Id)

		// Associate mesos Slave ID to the instance.
		i.SlaveId = found.SlaveId.GetValue()
		model.Instances(ctx).Update(i)
	}
	_, err = driver.LaunchTasks(acceptIDs, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	if err != nil {
		log.WithError(err).Error("Faild to response LaunchTasks.")
	}

	exists := func(s []*mesos.OfferID, i *mesos.OfferID) bool {
		for _, o := range s {
			if o.GetValue() == i.GetValue() {
				return true
			}
		}
		return false
	}
	for _, offer := range offers {
		if !exists(acceptIDs, offer.GetId()) {
			_, err := driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
			if err != nil {
				log.WithError(err).Error("Failed to response DeclineOffer.")
			}
		}
	}
	return nil
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

func startAPIServer(ctx context.Context, laddr string, zkAddr string, driver sched.SchedulerDriver) *api.APIServer {
	lis, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalln("Faild to bind address for gRPC API: ", laddr)
	}
	log.Println("Listening gRPC API on: ", laddr)
	s := api.NewAPIServer(zkAddr, driver, ctx)
	go s.Serve(lis)
	return s
}

func Run(ctx context.Context, listenAddr string, apiListenAddr string, mesosMasterAddr string, zkAddr string) {
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
		Scheduler:      newVDCScheduler(ctx, listenAddr, zkAddr),
		Framework:      FrameworkInfo,
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

	apiServer := startAPIServer(ctx, apiListenAddr, zkAddr, driver)
	defer func() {
		apiServer.GracefulStop()
	}()

	if stat, err := driver.Run(); err != nil {
		log.Printf("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
	}
}
