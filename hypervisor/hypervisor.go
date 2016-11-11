// +build linux

package hypervisor

import (
        "time"
        "log"

        lxc "gopkg.in/lxc/go-lxc.v2"
)

var lxcpath string = lxc.DefaultConfigPath()
var name string = "lxc-test"

func NewLxcContainer(imageName string, hostName string) {

        c, err := lxc.NewContainer(name, lxcpath)
        if err != nil {
                log.Println("ERROR: %s\n", err)
        }

        log.Println("Creating lxc-container...\n")
        //if verbose {
          //      c.SetVerbosity(lxc.Verbose)
        //}

        options := lxc.TemplateOptions{
                Template:             download,
                Distro:               ubuntu,
                Release:              trusty,
                Arch:                 amd64,
                FlushCache:           false,
                DisableGPGValidation: false,
        }

        if err := c.Create(options); err != nil {
                log.Println("ERROR: %s\n", err)
        }
}

func DestroyLxcContainer() {

        c, err := lxc.NewContainer(name, lxcpath)
        if err != nil {
                log.Println("ERROR: %s\n", err)
        }

        log.Println("Destroying lxc-container..\n")
        if err := c.Destroy(); err != nil {
                log.Println("ERROR: %s\n", err)
        }
}

func StartLxcContainer() {

        c, err := lxc.NewContainer(name, lxcpath)
        if err != nil {
                log.Println("ERROR: %s\n", err)
        }

        log.Println("Starting lxc-container...\n")
        if err := c.Start(); err != nil {
                log.Println("ERROR: %s\n", err)
        }

        log.Println("Waiting for lxc-container to start networking\n")
        if _, err := c.WaitIPAddresses(5 * time.Second); err != nil {
                log.Println("ERROR: %s\n", err.Error())
        }
}

func StopLxcContainer() {

        c, err := lxc.NewContainer(name, lxcpath)
        if err != nil {
                log.Println("ERROR: %s\n", err.Error())
        }

        log.Println("Stopping lxc-container..\n")
        if err := c.Stop(); err != nil {
                log.Println("ERROR: %s\n", err.Error())
        }
}


