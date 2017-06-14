// +build ignore

package main

// Build script for OpenVDC

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ProtocVersion = "libprotoc 3.2.0"
var verbose = true

// Similar to "$()" or "``" in sh
func cmd(cmd string, args ...string) string {
	if verbose {
		fmt.Println(cmd, strings.Join(args, " "))
	}
	run := exec.Command(cmd, args...)
	outbuf := new(bytes.Buffer)
	run.Stdout = outbuf
	run.Stderr = os.Stderr
	go func() {
		io.Copy(os.Stdout, outbuf)
	}()
	if err := run.Run(); err != nil {
		log.Fatalf("ERROR: %s: %s %s", err, cmd, strings.Join(args, " "))
	}
	return strings.TrimSpace(outbuf.String())
}

func pipe(cmd *exec.Cmd, cb func(line string)) {
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	s := bufio.NewScanner(out)
	for s.Scan() {
		cb(s.Text())
	}
	if s.Err() != nil {
		log.Fatal(s.Err())
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func if_not_exists(cmd string, cb func()) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		cb()
	}
}

func determineGHRef() string {
	var branchName string
	var exists bool
	branchName, exists = os.LookupEnv("APPVEYOR_REPO_BRANCH")
	if exists {
		return branchName
	}
	SCHEMA_LAST_COMMIT, exists := os.LookupEnv("SCHEMA_LAST_COMMIT")
	if !exists {
		SCHEMA_LAST_COMMIT = cmd("git", "log", "-n", "1", "--pretty=format:%H", "--", "schema/", "registry/schema.bindata.go")
	}

	found := false
	// Equivalent: (git rev-list origin/master | grep "${SCHEMA_LAST_COMMIT}") > /dev/null
	pipe(exec.Command("git", "rev-list", "origin/master"), func(l string) {
		if SCHEMA_LAST_COMMIT == l {
			found = true
		}
	})
	if found {
		// Found no changes for resource template/schema on HEAD.
		// so that set preference to the master branch.
		return "master"
	} else {
		// Found resource template/schema changes on this HEAD. Switch the default reference branch.
		// Check if $GIT_BRANCH/$BRANCH_NAME has something once in case of running in Jenkins.
		if branchName, exists = os.LookupEnv("BRANCH_NAME"); exists {
			return branchName
		}
		if branchName, exists = os.LookupEnv("GIT_BRANCH"); exists {
			return branchName
		}
	}
	return cmd("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func main() {
	// Apppend $GOBIN and $GOPATH/bin to $PATH
	if dir, ok := os.LookupEnv("GOBIN"); ok {
		os.Setenv("PATH", os.Getenv("PATH") + string(filepath.ListSeparator) + dir)
	}
	for _, dir := range filepath.SplitList(os.Getenv("GOPATH")) {
		os.Setenv("PATH", os.Getenv("PATH") + string(filepath.ListSeparator) + filepath.Join(dir, "bin"))
	}
	var with_gogen bool
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of build.go:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Environment Variables:
	VERSION: VERSION number to embed (default: dev)
	SHA:	SHA commit ID to embed (default from git rev-parse HEAD)`)
	}
	flag.BoolVar(&with_gogen, "with-gogen", false, "Run go generate for .proto and go-bindata")
	flag.Parse()

	var exists bool
	VERSION, exists := os.LookupEnv("VERSION")
	if !exists {
		VERSION = "dev"
	}

	SHA, exists := os.LookupEnv("SHA")
	if !exists {
		SHA = cmd("git", "rev-parse", "--verify", "HEAD")
	}
	BUILDDATE := time.Now().UTC().Format(time.RFC3339)

	GOVERSION := cmd("go", "version")
	LDFLAGS := fmt.Sprintf(
		"-X '%[1]s.Version=%[2]s' -X '%[1]s.Sha=%[3]s' -X '%[1]s.Builddate=%[4]s' -X '%[1]s.Goversion=%[5]s'",
		"github.com/axsh/openvdc/cmd",
		VERSION,
		SHA,
		BUILDDATE,
		GOVERSION,
	)

	LDFLAGS += fmt.Sprintf(
		" -X '%[1]s.GithubDefaultRef=%[2]s'",
		"github.com/axsh/openvdc/registry",
		determineGHRef(),
	)

	if_not_exists("go-bindata", func() {
		cmd("go", "get", "-u", "github.com/jteeuwen/go-bindata/...")
	})
	if_not_exists("go-bindata-assetfs", func() {
		cmd("go", "get", "-u", "github.com/elazarl/go-bindata-assetfs/...")
	})

	if with_gogen {
		if_not_exists("protoc", func() {
			log.Fatal("Unable to find protoc. Download pre-compiled binary from https://github.com/google/protobuf/releases")
		})
		// Check protoc version
		ver, err := exec.Command("protoc", "--version").Output()
		if err != nil {
			log.Fatalf("Failed to check protoc version: %v", err)
		}
		ver_str := strings.TrimSpace(string(ver))
		if ver_str != ProtocVersion {
			log.Fatalf("Unexpected protoc version: %s (expected: %s)", ver_str, ProtocVersion)
		}
		if_not_exists("protoc-gen-go", func() {
			cmd("go", "get", "-u", "-v", "github.com/golang/protobuf/protoc-gen-go")
		})
		cmd("go", "generate", "-v", "./api/...", "./model", "./registry")
	}

	if_not_exists("govendor", func() {
		cmd("go", "get", "-u", "github.com/kardianos/govendor")
	})
	cmd("govendor", "sync")

	// Build main binaries
	cmd("go", "build", "-i", "./vendor/...")
	cmd("go", "build", "-ldflags", LDFLAGS, "-v", "./cmd/openvdc")
	cmd("go", "build", "-ldflags", LDFLAGS+
		" -X 'main.HostRsaKeyPath=/etc/openvdc/ssh/host_rsa_key'" +
		" -X 'main.HostEcdsaKeyPath=/etc/openvdc/ssh/host_ecdsa_key'" +
		" -X 'main.HostEd25519KeyPath=/etc/openvdc/ssh/host_ed25519_key'" +
		" -X 'main.DefaultConfPath=/etc/openvdc/executor.toml'", "-v", "./cmd/openvdc-executor")
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.DefaultConfPath=/etc/openvdc/scheduler.toml'", "-v", "./cmd/openvdc-scheduler")

	//Build lxc-template
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.lxcPath=/usr/share/lxc/'", "-v", "-o", "./lxc-openvdc", "./cmd/lxc-openvdc/")

	//Build qemu-ifup/qemi-ifdown
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.DefaultConfPath=/etc/openvdc/executor.toml'", "-v", "-o", "./qemu-ifup", "./cmd/qemu-ifup")
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.DefaultConfPath=/etc/openvdc/executor.toml'", "-v", "-o", "./qemu-ifdown", "./cmd/qemu-ifdown")

	// Build Acceptance Test binary
	os.Chdir("./ci/citest/acceptance-test/tests")
	cmd("govendor", "sync")
	cmd("go", "generate", "-v", "-tags=acceptance", ".")
	cmd("go", "test", "-tags=acceptance", "-c", "-o", "openvdc-acceptance-test")
}
