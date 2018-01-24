# Computing resources

OpenVDC uses the mesos frameworks resource handling to keep track of available
cpu/memory on each executor. By default the values are fetched from the
physical resources available on the host running the executor.

Depending on the hypervisor driver being used, we need to manually define the
resources available.

* LXC are simple containers and does not fully virtualize memory/cpu.
Resource configuration is not required unless there is a reason to restrict the
resources offered.

* Qemu assigns virtual cpu cores to instances booted. Here we can overwrite the
default cpus value to a larger number since we are not using physical cores for
our instances. Memory is calculated based on the available memory when the agent
is started and should be left alone as it is a physical resource.

* ESXi requires the executor to run on a remote host. In this case all computing
resources needs to be set manually as the mesos agent is only able to fetch local
resources. Like with qemu cpus are virtual resources and can be overwritten to
fit requirements. Memory should be set to match the available memory on the
ESXi host.


To overwrite the default values create `/etc/mesos-slave/resources` before
starting the mesos-slave agent

Following is a basic example

```
[
  {
    "name": "cpus",
    "type": "SCALAR",
    "scalar": {
      "value": 20
    }
  },
  {
    "name": "mem",
    "type": "SCALAR",
    "scalar": {
      "value": 20480
    }
  }
]
```

reference http://mesos.apache.org/documentation/latest/attributes-resources/
