package scheduler

import (
	"fmt"
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/gogo/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/auth/sasl"
	_ "github.com/mesos/mesos-go/detector/zoo"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/pkg/errors"
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

	checkAgents(offers, ctx)

	if sched.tasksLaunched == 0 {
		sched.CheckForCrashedNodes(offers, ctx)
	}

	disconnected := getDisconnectedInstances(offers, ctx, driver)

	if len(disconnected) > 0 {
		sched.InstancesRelaunching(driver, offers, ctx, disconnected)
	} else {
		sched.InstancesQueued(driver, offers, ctx)
	}

	return nil
}

func (sched *VDCScheduler) CheckForCrashedNodes(offers []*mesos.Offer, ctx context.Context) error {
	for _, offer := range offers {
		node, err := model.Nodes(ctx).FindByAgentID(getAgentID(offer))
		if err != nil {
			log.WithError(err).Error("Failed to get node")
		}

		if node == nil {
			continue
		}

		instances, err := model.Instances(ctx).FilterByAgentMesosID(node.GetAgentMesosID())

		if err != nil {
			return errors.Wrapf(err, "Failed to retrieve instances.")
		}

		foundCrashedNode := false

	CheckInstances:
		for _, instance := range instances {
			if instance.GetLastState().State != model.InstanceState_REGISTERED &&
				instance.GetLastState().State != model.InstanceState_QUEUED &&
				instance.GetLastState().State != model.InstanceState_TERMINATED &&
				instance.GetAutoRecovery() == true {

				disconnectedAgent, err := model.CrashedNodes(ctx).FindByAgentMesosID(*offer.SlaveId.Value)

				if err != nil {
					return errors.Wrapf(err, "Failed to check if crashed node existed.")
				}

				if disconnectedAgent == nil || disconnectedAgent.GetReconnected() == true {
					foundCrashedNode = true
					break CheckInstances
				}
			}
		}

		if foundCrashedNode == true {

			agentID := getAgentID(offer)

			err := model.Nodes(ctx).UpdateAgentMesosID(agentID, *offer.SlaveId.Value)
			if err != nil {
				log.WithError(err).Error("Failed to update node agentMesosID. node: '%s'", agentID)
			}

			err = model.CrashedNodes(ctx).Add(&model.CrashedNode{
				Agentid:      agentID,
				Agentmesosid: *offer.SlaveId.Value,
				Reconnected:  false,
			})

			if err != nil {
				return errors.Wrapf(err, "Failed to add '%s' to lost of crashed agents.", agentID)
			}

			log.Infoln("Added '%s' to list of crashed agents", agentID)

			instances, err := model.Instances(ctx).FilterByAgentMesosID(*offer.SlaveId.Value)

			if err != nil {
				return errors.Wrapf(err, "Failed to retrieve instances.")
			}

			if len(instances) > 0 {
				for _, instance := range instances {
					err := model.Instances(ctx).UpdateConnectionStatus(instance.GetId(), model.ConnectionStatus_NOT_CONNECTED)
					if err != nil {
						return errors.Wrapf(err, "Failed to update instance ConnectionStatus. instance: '%s' ConnectionStatus: '%s'", instance.GetId(), model.ConnectionStatus_NOT_CONNECTED)
					}
				}
			}
		}
	}

	return nil
}

func getDisconnectedInstances(offers []*mesos.Offer, ctx context.Context, driver sched.SchedulerDriver) []*model.Instance {

	disconnectedInstances := []*model.Instance{}

	for _, offer := range offers {

		agentID := getAgentID(offer)
		disconnectedAgent, err := model.CrashedNodes(ctx).FindByAgentID(agentID)

		if err != nil {
			log.WithError(err).Error("Failed to retrieve crashed node.")
		}

		if disconnectedAgent != nil {
			if disconnectedAgent.GetReconnected() == false {
				instances, err := model.Instances(ctx).FilterByAgentMesosID(disconnectedAgent.GetAgentMesosID())

				if err != nil {
					log.WithError(err).Error("Failed to retrieve disconnected instances.")
				}

				if len(instances) > 0 {
					for _, instance := range instances {
						connStatus := instance.GetConnectionStatus()
						autoRecovery := instance.GetAutoRecovery()

						if connStatus.Status == model.ConnectionStatus_NOT_CONNECTED && autoRecovery == true {
							instance.SlaveId = offer.SlaveId.GetValue()
							model.Instances(ctx).Update(instance)

							log.Infoln(fmt.Sprintf("Added instance %s to relaunch-queue.", instance.GetId()))
							disconnectedInstances = append(disconnectedInstances, instance)
						}
					}
				}

				if len(disconnectedInstances) == 0 {
					model.CrashedNodes(ctx).SetReconnected(disconnectedAgent)
					agentID := disconnectedAgent.GetAgentID()
					log.Infoln(fmt.Sprintf("Node '%s' reconnected.", agentID))

					err := model.Nodes(ctx).UpdateAgentMesosID(agentID, *offer.SlaveId.Value)
					if err != nil {
						log.WithError(err).Error("Failed to update node agentMesosID. node: '%s'", agentID)
					}
				}
			}
		}
	}
	return disconnectedInstances
}

func (sched *VDCScheduler) InstancesRelaunching(driver sched.SchedulerDriver, offers []*mesos.Offer, ctx context.Context, relaunchQueued []*model.Instance) error {

	tasks := []*mesos.TaskInfo{}
	acceptIDs := []*mesos.OfferID{}

	for _, instance := range relaunchQueued {
		for _, offer := range offers {
			if instance.SlaveId != offer.SlaveId.GetValue() {
				continue
			}

			alreadyAdded := false

			for i, _ := range acceptIDs {
				if acceptIDs[i] == offer.Id {
					alreadyAdded = true
				}
			}

			if alreadyAdded != true {
				hypervisorName := getHypervisorName(offer)
				model.Instances(ctx).UpdateConnectionStatus(instance.GetId(), model.ConnectionStatus_CONNECTED)
				task := sched.NewTask(instance, util.NewSlaveID(instance.SlaveId), ctx, sched.NewExecutor(hypervisorName))
				tasks = append(tasks, task)
				acceptIDs = append(acceptIDs, offer.Id)
			}
		}
	}

	sched.LaunchTasks(driver, tasks, acceptIDs, offers)
	sched.DeclineUnusedOffers(driver, offers, acceptIDs)

	return nil
}

func (sched *VDCScheduler) InstancesQueued(driver sched.SchedulerDriver, offers []*mesos.Offer, ctx context.Context) error {
	queued, err := model.Instances(ctx).FilterByState(model.InstanceState_QUEUED)
	if err != nil {
		return errors.WithStack(err)
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

	tasks := []*mesos.TaskInfo{}
	acceptIDs := []*mesos.OfferID{}

	for _, i := range queued {
		found := findMatching(i, offers, ctx)
		for i, _ := range acceptIDs {
			if acceptIDs[i] == found.Id {
				found = nil
			}
		}

		if found == nil {
			continue
		}

		hypervisorName := getHypervisorName(found)
		log.WithFields(log.Fields{
			"instance_id": i.GetId(),
			"hypervisor":  hypervisorName,
		}).Info("Found matching offer")

		task := sched.NewTask(i, found.SlaveId, ctx, sched.NewExecutor(hypervisorName))
		tasks = append(tasks, task)
		acceptIDs = append(acceptIDs, found.Id)

		i.SlaveId = found.SlaveId.GetValue()
		model.Instances(ctx).Update(i)
	}

	sched.LaunchTasks(driver, tasks, acceptIDs, offers)
	sched.DeclineUnusedOffers(driver, offers, acceptIDs)

	return nil
}

func (sched *VDCScheduler) NewTask(i *model.Instance, slaveID *mesos.SlaveID, ctx context.Context, executor *mesos.ExecutorInfo) *mesos.TaskInfo {
	taskId := util.NewTaskID(i.GetId())
	task := &mesos.TaskInfo{
		Name:     proto.String("VDC" + "_" + taskId.GetValue()),
		TaskId:   taskId,
		SlaveId:  slaveID,
		Data:     []byte("instance_id=" + i.GetId()),
		Executor: executor,
		Resources: []*mesos.Resource{
			util.NewScalarResource("cpus", float64(i.GetTemplate().GetLxc().GetVcpu())),
			util.NewScalarResource("mem", float64(i.GetTemplate().GetLxc().GetMemoryGb()*1024)),
		},
	}
	return task
}

func (sched *VDCScheduler) NewExecutor(hypervisorName string) *mesos.ExecutorInfo {
	executor := &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID(fmt.Sprintf("vdc-hypervisor-%s", hypervisorName)),
		Name:       proto.String("VDC Executor"),
		Command: &mesos.CommandInfo{
			Value: proto.String(fmt.Sprintf("%s --hypervisor=%s --zk=%s",
				ExecutorPath, hypervisorName, sched.zkAddr.String())),
		},
	}
	return executor
}

func (sched *VDCScheduler) LaunchTasks(driver sched.SchedulerDriver, tasks []*mesos.TaskInfo, acceptIDs []*mesos.OfferID, offers []*mesos.Offer) error {
	_, err := driver.LaunchTasks(acceptIDs, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	if err != nil {
		return errors.Wrapf(err, "Failed to launch tasks. Tasks: %s", tasks)
	}
	sched.tasksLaunched++

	return nil
}

func findMatching(i *model.Instance, offers []*mesos.Offer, ctx context.Context) *mesos.Offer {
	for _, offer := range offers {
		log := log.WithField("agent", offer.SlaveId.String())
		var agentAttrs struct {
			Hypervisor string   // Required
			NodeGroups []string // Optional
		} // Read and validate attribute entries from agent offer.
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
		default:
			log.Warnf("Unknown template type: %T", t)
		}
	}
	return nil
}

func (sched *VDCScheduler) DeclineUnusedOffers(driver sched.SchedulerDriver, offers []*mesos.Offer, acceptIDs []*mesos.OfferID) {
	exists := func(s []*mesos.OfferID, i *mesos.OfferID) bool {
		for _, o := range s {
			if o.GetValue() == i.GetValue() {
				return true
			}
		}
		return false
	}

	for _, offer := range offers {
		if !exists(acceptIDs, offer.Id) {
			_, err := driver.DeclineOffer(offer.Id, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
			if err != nil {
				log.WithError(err).Error("Failed to decline offer.")
			}
		}
	}
}

func checkAgents(offers []*mesos.Offer, ctx context.Context) error {
	for _, offer := range offers {
		slaveID := offer.SlaveId.GetValue()
		agentID := getAgentID(offer)

		err := model.Nodes(ctx).Add(&model.AgentNode{
			Agentmesosid: slaveID,
			Agentid:      agentID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getAgentID(offer *mesos.Offer) string {
	for _, attr := range offer.Attributes {
		if attr.GetName() == "openvdc-node-id" &&
			attr.GetType() == mesos.Value_TEXT {
			return attr.GetText().GetValue()
		}
	}
	return ""
}

func getHypervisorName(offer *mesos.Offer) string {
	for _, attr := range offer.Attributes {
		if attr.GetName() == "hypervisor" &&
			attr.GetType() == mesos.Value_TEXT {
			return attr.GetText().GetValue()
		}
	}
	return ""
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
		driver.ReviveOffers()
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

	agentMesosID := *sid.Value

	ctx, err := model.Connect(context.Background(), sched.zkAddr)
	if err != nil {
		log.WithError(err).Error("Failed model.Connect")
	}

	instances, err := model.Instances(ctx).FilterByAgentMesosID(agentMesosID)

	if err != nil {
		log.WithError(err).Error("Failed to retrieve instances.")
	}

	if len(instances) > 0 {
		for _, instance := range instances {
			err := model.Instances(ctx).UpdateConnectionStatus(instance.GetId(), model.ConnectionStatus_NOT_CONNECTED)
			if err != nil {
				log.WithError(err).Error("Failed to update instance ConnectionStatus. instance: '%s' ConnectionStatus: '%s'", instance.GetId(), model.ConnectionStatus_NOT_CONNECTED)
			}
		}
	}

	res, err := model.Nodes(ctx).FindByAgentMesosID(agentMesosID)
	if err != nil {
		log.WithError(err).Error("Failed to fetch agent nodes")
	}

	if res != nil {

		agentID := res.Agentid

		err = model.CrashedNodes(ctx).Add(&model.CrashedNode{
			Agentid:      agentID,
			Agentmesosid: agentMesosID,
			Reconnected:  false,
		})

		if err != nil {
			log.WithError(err).Error("Failed to add '%s' to list of crashed agents", agentID)
		}

		log.Infoln("Added '%s' to list of crashed agents", agentID)
	}
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
