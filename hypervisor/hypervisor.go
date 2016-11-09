// +build linux

package hypervisor

import (
        "flag"
        "time"
        "log"

        lxc "gopkg.in/lxc/go-lxc.v2"
)

var (
        slowTasks  = flag.Bool("slow_tasks", false, "")
        imageName  = flag.String("imageName", "", "")
        hostName   = flag.String("hostName", "", "")
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

func NewLxcContainer() {

      c, err := lxc.NewContainer(name, lxcpath)
        if err != nil {
                log.Println("ERROR: %s\n", err)
        }

        log.Println("Creating lxc-container...\n")
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


