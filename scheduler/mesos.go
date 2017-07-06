package scheduler

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"strings"

	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/auth/sasl"
	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"
)

var ExecutorPath string

type SchedulerSettings struct {
	Name            string
	ID              string
	FailoverTimeout float64
	ExecutorPath    string
}

type VDCScheduler struct {
	tasksLaunched int
	tasksFinished int
	tasksErrored  int
	totalTasks    int
	listenAddr    string
	zkAddr        backend.ZkEndpoint
	ctx           context.Context
}

func newVDCScheduler(ctx context.Context, listenAddr string, zkAddr backend.ZkEndpoint) *VDCScheduler {
	return &VDCScheduler{
		listenAddr: listenAddr,
		zkAddr:     zkAddr,
		ctx:        ctx,
	}
}

func (sched *VDCScheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Println("Framework Registered with Master ", masterInfo)
	node := &model.SchedulerNode{
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

	_, err := driver.ReconcileTasks([]*mesos.TaskStatus{})

	if err != nil {
		log.Errorln("Failed to reconcile tasks: %v", err)
	}
}

func (sched *VDCScheduler) Disconnected(sched.SchedulerDriver) {
	log.Println("disconnected from master")
}

func (sched *VDCScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	log := log.WithFields(log.Fields{"offers": len(offers)})

	ctx, err := model.Connect(context.Background(), sched.zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed to connect to datasource")
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

	findMatching := func(i *model.Instance) *mesos.Offer {
		log := log.WithField("instance_id", i.GetId())
		for _, offer := range offers {
			log := log.WithField("agent", offer.SlaveId.String())
			var agentAttrs struct {
				Hypervisor string   // Required
				NodeGroups []string // Optional
			}
			// Read and validate attribute entries from agent offer.
			for _, attr := range offer.Attributes {
				switch attr.GetName() {
				case "hypervisor":
					if attr.GetType() != mesos.Value_TEXT {
						log.Error("Invalid value type for 'hypervisor' attribute")
						break
					}
					agentAttrs.Hypervisor = attr.GetText().GetValue()

				case "node-groups":
					if attr.GetType() == mesos.Value_TEXT {
						if attr.GetText().GetValue() == "" {
							log.Error("'node-groups' attribute must be non-empty string")
							break
						}
						agentAttrs.NodeGroups = strings.Split(attr.GetText().GetValue(), ",")
					} else {
						log.Errorf("Invalid value type for 'bridge' attribute: %s", attr.GetText())
						break
					}
				default:
					log.Warnf("Found unsupported attribute: %s", attr.GetName())
				}
			}

			if agentAttrs.Hypervisor == "" {
				log.Error("Required attributes are not advertised from agent")
				continue
			}

			// TODO: Avoid type switch to find template types.
			switch t := i.GetTemplate().GetItem(); t.(type) {
			case *model.Template_Lxc:
				if agentAttrs.Hypervisor == "lxc" {
					lxc := i.GetTemplate().GetLxc()
					if !model.IsMatchingNodeGroups(lxc, agentAttrs.NodeGroups) {
						return nil
					}
					return offer
				}
			case *model.Template_Null:
				if agentAttrs.Hypervisor == "null" {
					return offer
				}
			case *model.Template_Esxi:
				if agentAttrs.Hypervisor == "esxi" {
					esxi := i.GetTemplate().GetEsxi()
					if !model.IsMatchingNodeGroups(esxi, agentAttrs.NodeGroups) {
						return nil
					}
					return offer
				}
			default:
				log.Warnf("Unknown template type: %T", t)
			}
		}
		return nil
	}

	tasks := []*mesos.TaskInfo{}
	acceptIDs := []*mesos.OfferID{}
	for _, i := range queued {
		found := findMatching(i)
		for i, _ := range acceptIDs {
			if acceptIDs[i] == found.Id {
				found = nil
			}
		}

		if found == nil {
			continue
		}

		hypervisorName := strings.TrimPrefix(i.GetTemplate().ResourceTemplate().ResourceName(), "vm/")
		log.WithFields(log.Fields{
			"instance_id": i.GetId(),
			"hypervisor":  hypervisorName,
		}).Info("Found matching offer")

		executor := &mesos.ExecutorInfo{
			ExecutorId: util.NewExecutorID(fmt.Sprintf("vdc-hypervisor-%s", hypervisorName)),
			Name:       proto.String("VDC Executor"),
			Command: &mesos.CommandInfo{
				Value: proto.String(fmt.Sprintf("%s --hypervisor=%s --zk=%s",
					ExecutorPath, hypervisorName, sched.zkAddr.String())),
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
				util.NewScalarResource("cpus", float64(i.GetTemplate().GetLxc().GetVcpu())),
				util.NewScalarResource("mem", float64(i.GetTemplate().GetLxc().GetMemoryGb()*1024)),
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
	log.Infoln("offer rescinded: %v", oid)
}

func (sched *VDCScheduler) FrameworkMessage(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, msg string) {
	log.Infoln("framework message from executor %q slave %q: %q", eid, sid, msg)
}

func (sched *VDCScheduler) SlaveLost(_ sched.SchedulerDriver, sid *mesos.SlaveID) {
	log.Errorln("slave lost: %v", sid)
}

func (sched *VDCScheduler) ExecutorLost(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, code int) {
	log.Errorln("executor %q lost on slave %q code %d", eid, sid, code)
}

func (sched *VDCScheduler) Error(_ sched.SchedulerDriver, err string) {
	log.Errorln("Scheduler received error: %v", err)
}

func NewMesosScheduler(ctx context.Context, listenAddr string, mesosMasterAddr string, zkAddr backend.ZkEndpoint, settings SchedulerSettings) (*sched.MesosSchedulerDriver, error) {
	cred := &mesos.Credential{
		Principal: proto.String(""),
		Secret:    proto.String(""),
	}

	cred = nil
	bindingAddrs, err := net.LookupIP(listenAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid listen address: %s", listenAddr)
	}

	ExecutorPath = settings.ExecutorPath

	FrameworkInfo := &mesos.FrameworkInfo{
		User:            proto.String(""),
		Name:            proto.String(settings.Name),
		FailoverTimeout: proto.Float64(settings.FailoverTimeout),
		Id:              util.NewFrameworkID(settings.ID),
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
		return nil, errors.Wrap(err, "Unable to create SchedulerDriver")
	}
	return driver, nil
}
