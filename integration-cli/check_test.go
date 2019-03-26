package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/docker/docker/test/fakestorage"
	"github.com/docker/docker/integration-cli/cli"
	"github.com/docker/docker/integration-cli/environment"
	"github.com/docker/docker/pkg/reexec"
	testdaemon "github.com/docker/docker/test/daemon"
	ienv "github.com/docker/docker/test/environment"
	"github.com/go-check/check"
)

const (
	// path to containerd's ctr binary
	ctrBinary = "ctr"

	// the docker daemon binary to use
	dockerdBinary = "dockerd"
)

var (
	testEnv *environment.Execution

	// the docker client binary to use
	dockerBinary = ""
)

func init() {
	var err error

	reexec.Init() // This is required for external graphdriver tests

	testEnv, err = environment.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	dockerBinary = testEnv.DockerBinary()
	err := ienv.EnsureFrozenImagesLinux(&testEnv.Execution)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testEnv.Print()
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	cli.SetTestEnvironment(testEnv)
	fakestorage.SetTestEnvironment(&testEnv.Execution)
	ienv.ProtectAll(t, &testEnv.Execution)
	check.TestingT(t)
}

func init() {
	check.Suite(&DockerSuite{})
}

type DockerSuite struct {
}

func (s *DockerSuite) OnTimeout(c *check.C) {
	if testEnv.IsRemoteDaemon() {
		return
	}
	dest := os.Getenv("DEST")
	if dest != "" {
		dest = "/run"
	}
	path := filepath.Join(os.Getenv("DEST"), "docker.pid")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		c.Fatalf("Failed to get daemon PID from %s\n", path)
	}

	rawPid, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		c.Fatalf("Failed to parse pid from %s: %s\n", path, err)
	}

	daemonPid := int(rawPid)
	if daemonPid > 0 {
		testdaemon.SignalDaemonDump(daemonPid)
	}
}

func (s *DockerSuite) TearDownTest(c *check.C) {
	testEnv.Clean(c)
}
