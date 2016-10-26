// +build linux

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	exec "github.com/mesos/mesos-go/executor"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/samuel/go-zookeeper/zk"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

var (
	slowTasks  = flag.Bool("slow_tasks", false, "")
	lxcpath    string
	template   string
	distro     string
	release    string
	arch       string
	name       string
	verbose    bool
	flush      bool
	validation bool
)

type VDCExecutor struct {
	tasksLaunched int
}

func newVDCExecutor() *VDCExecutor {
	return &VDCExecutor{tasksLaunched: 0}
}

func (exec *VDCExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.SlaveInfo) {
	fmt.Println("Registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *VDCExecutor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.SlaveInfo) {
	fmt.Println("Re-registered Executor on slave ", slaveInfo.GetHostname())
}

func (exec *VDCExecutor) Disconnected(driver exec.ExecutorDriver) {
	fmt.Println("Executor disconnected.")
}

func zkConnect(ip string) *zk.Conn {
	c, _, err := zk.Connect([]string{ip}, time.Second)

	if err != nil {
		fmt.Println("VDCExecutor failed connecting to Zookeeper: ", err)
	}

	return c
}

func zkGetData(c *zk.Conn, dir string) []byte {
	data, stat, err := c.Get(dir)

	if err != nil {
		fmt.Println("VDCExecutor failed getting data from Zookeeper: ", err)
	}

	fmt.Println(stat)

	return data[:]
}

func zkSendData(c *zk.Conn, dir string, data string) {
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)

	path, err := c.Create(dir, []byte(data), flags, acl)

	if err != nil {
		fmt.Println("VDCExecutor failed sending data to Zookeeper: ", err)
	}

	fmt.Println("Sent: ", data, "to ", dir)
	fmt.Println(path)
}

func testZkConnection(ip string) {

	c := zkConnect(ip)
	zkSendData(c, "/02", "This is a test")
	data := []byte(zkGetData(c, "/02"))
}

func newLxcContainer() *lxc.Container {

	c, err := lxc.NewContainer(name, lxcpath)
	if err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}

	fmt.Println("Creating lxc-container...\n")
	if verbose {
		c.SetVerbosity(lxc.Verbose)
	}

	options := lxc.TemplateOptions{
		Template:             template,
		Distro:               distro,
		Release:              release,
		Arch:                 arch,
		FlushCache:           flush,
		DisableGPGValidation: validation,
	}

	if err := c.Create(options); err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}

	return c
}

func destroyLxcContainer(c *lxc.Container) {

	fmt.Println("Destroying lxc-container...\n")
	if err := c.Destroy(); err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}
}

func startLxcContainer(c *lxc.Container) {

	fmt.Println("Starting lxc-container...\n")
	if err := c.Start(); err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}

	fmt.Println("Waiting for lxc-container to start networking\n")
	if _, err := c.WaitIPAddresses(5 * time.Second); err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}
}

func stopLxcContainer(c *lxc.Container) {

	fmt.Println("Stopping lxc-container...\n")
	if err := c.Stop(); err != nil {
		fmt.Println("ERROR: %s\n", err.Error())
	}
}

func trimName(untrimmedName string) string {
        limit := "_"
        trimmedName := strings.Split(untrimmedName, limit)[0]

        return trimmedName
}

func newTask(taskName string) {

	trimmedTaskName := trimName(taskName)

        switch trimmedTaskName {
                case "lxc-create":
                        //lxc := newLxcContainer()
                case "lxc-start":
                        //startLxcContainer(lxc)
                case "lxc-stop":
                        //stopLxcContainer(lxc)
                case "lxc-destroy":
                        //destroyLxcContainer(lxc)
                default:
                        fmt.Println("ERROR: Taskname unrecognized")
        }
}


func (exec *VDCExecutor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	fmt.Println("Launching task", taskInfo.GetName(), "with command", taskInfo.Command.GetValue())

	runStatus := &mesos.TaskStatus{
		TaskId: taskInfo.GetTaskId(),
		State:  mesos.TaskState_TASK_RUNNING.Enum(),
	}
	_, err := driver.SendStatusUpdate(runStatus)
	if err != nil {
		fmt.Println("Got error", err)
	}

	exec.tasksLaunched++
	fmt.Println("Total tasks launched ", exec.tasksLaunched)

	newTask(taskInfo.GetName())


	finishTask := func() {
		fmt.Println("Finishing task", taskInfo.GetName())
		finStatus := &mesos.TaskStatus{
			TaskId: taskInfo.GetTaskId(),
			State:  mesos.TaskState_TASK_FINISHED.Enum(),
		}
		if _, err := driver.SendStatusUpdate(finStatus); err != nil {
			fmt.Println("error sending FINISHED", err)
		}
		fmt.Println("Task finished", taskInfo.GetName())
	}
	if *slowTasks {
		starting := &mesos.TaskStatus{
			TaskId: taskInfo.GetTaskId(),
			State:  mesos.TaskState_TASK_STARTING.Enum(),
		}
		if _, err := driver.SendStatusUpdate(starting); err != nil {
			fmt.Println("error sending STARTING", err)
		}
		delay := time.Duration(rand.Intn(90)+10) * time.Second
		go func() {
			time.Sleep(delay)
			finishTask()
		}()
	} else {
		finishTask()
	}
}

func (exec *VDCExecutor) KillTask(driver exec.ExecutorDriver, taskID *mesos.TaskID) {
	fmt.Println("Kill task")
}

func (exec *VDCExecutor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {
	fmt.Println("Got framework message: ", msg)
}

func (exec *VDCExecutor) Shutdown(driver exec.ExecutorDriver) {
	fmt.Println("Shutting down the executor")
}

func (exec *VDCExecutor) Error(driver exec.ExecutorDriver, err string) {
	fmt.Println("Got error message:", err)
}

func init() {
	flag.StringVar(&lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	flag.StringVar(&template, "template", "download", "Template to use")
	flag.StringVar(&distro, "distro", "ubuntu", "Template to use")
	flag.StringVar(&release, "release", "trusty", "Template to use")
	flag.StringVar(&arch, "arch", "amd64", "Template to use")
	flag.StringVar(&name, "name", "test", "Name of the container")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&flush, "flush", false, "Flush the cache")
	flag.BoolVar(&validation, "validation", false, "GPG validation")
	flag.Parse()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Starting VDC Executor (Go)")

	dconfig := exec.DriverConfig{
		Executor: newVDCExecutor(),
	}
	driver, err := exec.NewMesosExecutorDriver(dconfig)

	if err != nil {
		fmt.Println("Unable to create a ExecutorDriver ", err.Error())
	}

	_, err = driver.Start()
	if err != nil {
		fmt.Println("Got error:", err)
		return
	}
	fmt.Println("Executor process has started and running.")

	_, err = driver.Join()
	if err != nil {
		fmt.Println("driver failed:", err)
	}
	fmt.Println("executor terminating")
}
