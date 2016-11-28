package scheduler

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"

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
}

func newVDCScheduler(listenAddr string, zkAddr string) *VDCScheduler {
	return &VDCScheduler{
		totalTasks: taskCount,
		listenAddr: listenAddr,
		zkAddr:     zkAddr,
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
	queued, err := model.Instances(ctx).FilterByState(model.Instance_QUEUED)
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

	findMatching := func(i *model.Instance) *mesos.Offer {
		for _, offer := range offers {
			var hypervisorAttr *mesos.Attribute
			for _, attr := range offer.Attributes {
				if attr.GetName() == "hypervisor" &&
					attr.GetType() == mesos.Value_TEXT {
					hypervisorAttr = attr
				}
			}

			if hypervisorAttr == nil {
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
			switch t := r.GetTemplate(); t.(type) {
			case *model.Resource_Lxc:
				if hypervisorAttr.GetText().GetValue() == "lxc" {
					return offer
				}
			case *model.Resource_Null:
				if hypervisorAttr.GetText().GetValue() == "null" {
					return offer
				}
			}
		}
		return nil
	}

	tasks := []*mesos.TaskInfo{}
	acceptIDs := []*mesos.OfferID{}
	for _, i := range queued {
		log.Info("Queued Instance: ", i.GetId())
		found := findMatching(i)
		if found == nil {
			continue
		}
		log.WithField("InstanceID", i.GetId()).Info("Found matching offer")

		executor := &mesos.ExecutorInfo{
			ExecutorId: util.NewExecutorID("vdc-hypervisor-null"),
			Name:       proto.String("VDC Executor"),
			Command: &mesos.CommandInfo{
				Value: proto.String(fmt.Sprintf("%s -logtostderr=true -hypervisor=null -zk=%s", ExecutorPath, sched.zkAddr)),
			},
		}

		taskId := util.NewTaskID(uuid.New())
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
	}
	_, err = driver.LaunchTasks(acceptIDs, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	if err != nil {
		log.WithError(err).Error("Faild to response LaunchTasks.")
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

func startAPIServer(laddr string, zkAddr string, driver sched.SchedulerDriver) *api.APIServer {
	lis, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatalln("Faild to bind address for gRPC API: ", laddr)
	}
	log.Println("Listening gRPC API on: ", laddr)
	s := api.NewAPIServer(zkAddr, driver)
	go s.Serve(lis)
	return s
}

func Run(listenAddr string, apiListenAddr string, mesosMasterAddr string, zkAddr string) {
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
		Scheduler:      newVDCScheduler(listenAddr, zkAddr),
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

	apiServer := startAPIServer(apiListenAddr, zkAddr, driver)
	defer func() {
		apiServer.GracefulStop()
	}()

	if stat, err := driver.Run(); err != nil {
		log.Printf("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
	}
}
