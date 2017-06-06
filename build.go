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
	"archive/zip"
	"net/http"
	"runtime"
)

const ProtocVersion = "libprotoc 3.2.0" // strings.Split(ProtocVersion, " ")[1] is used below, so be careful changing this value...
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

func removeExt(s string) string {
	return s[:len(s)-len(filepath.Ext(s))]
}

func unzip(src, dest, fileToExtract string) error {
	r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
       rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            os.MkdirAll(filepath.Dir(path), f.Mode())
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
			if strings.Contains((*f).Name, fileToExtract) {
        err := extractAndWriteFile(f)
        if err != nil {
					return err
        }
				return nil
			}
    }

    return nil
}

func downloadFromUrl(url string) string {
	tokens := strings.Split(url, "/")
	fileName := filepath.Join(os.Getenv("GOPATH"), tokens[len(tokens)-1])
	fmt.Println("Downloading", url, "to", fileName)

	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return ""
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return ""
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return ""
	}

	fmt.Println(n, "bytes downloaded.")
	return fileName
}

func join(separator string, args ...string) string {
	return strings.Join(args, separator)
}

func installProtoc() {
		if_not_exists("protoc", func() { // since we require a specific version anyway, wouldn't it be better to include this binary in the initial install filesystem?
			const GOARCH string = runtime.GOARCH
			const GOOS string = runtime.GOOS
			if GOOS == "linux" && GOARCH == "amd64" { // check os and arch. If 64-bit linux, download binary. Other cases should be added as we start supporting other oses/architectures.
				protocVersionNumber := strings.Split(ProtocVersion, " ")[1]
				filename := downloadFromUrl(join("", "https://github.com/google/protobuf/releases/download/v", protocVersionNumber, "/protoc-", protocVersionNumber, "-linux-x86_64.zip"))
				dirname := removeExt(filename)
				log.Printf("Unzipping %s to %s", filename, dirname)
				if err := unzip(filename, dirname, "protoc"); err != nil {
					log.Println("Error unzipping file.")
				}
				GOPATH, exists := os.LookupEnv("GOPATH")
				if !exists {
					log.Println("GOPATH is not set. If protoc installation fails, please set the GOPATH environment variable and try again.")
				}
				log.Printf("Moving %s/bin/protoc to %s/bin/protoc", dirname, GOPATH)
				if err := os.Rename(filepath.Join(dirname,"bin/protoc"), filepath.Join(GOPATH, "bin/protoc")); err != nil{
					log.Println(err)
				}
				log.Printf("Removing: %s", filename)
				if err:= os.Remove(filename); err != nil{
					log.Println(err)
				}
				log.Printf("Removing: %s", dirname)
				if err:= os.RemoveAll(dirname); err != nil{
					log.Println(err)
				}
			} else {
			log.Fatalf("Unable to find protoc. Download a pre-compiled binary (version %s) for your system from https://github.com/google/protobuf/releases/tag/v3.2.0 to a suitable location found in your environment's PATH variable.", ProtocVersion)
			}
		})
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

	GOVERSION := runtime.Version()
	log.Println(GOVERSION)
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

	if with_gogen {		// Check protoc version
		installProtoc()
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
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.DefaultConfPath=/etc/openvdc/executor.toml'", "-v", "./cmd/openvdc-executor")
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.DefaultConfPath=/etc/openvdc/scheduler.toml'", "-v", "./cmd/openvdc-scheduler")

	//Build lxc-template
	cmd("go", "build", "-ldflags", LDFLAGS+"-X 'main.lxcPath=/usr/share/lxc/'", "-v", "-o", "./lxc-openvdc", "./cmd/lxc-openvdc/")

	// Build Acceptance Test binary
	os.Chdir("./ci/citest/acceptance-test/tests")
	cmd("govendor", "sync")
	cmd("go", "generate", "-v", "-tags=acceptance", ".")
	cmd("go", "test", "-tags=acceptance", "-c", "-o", "openvdc-acceptance-test")
}
