package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestRmContainerWithRemovedVolume(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, SameHostDaemon)

	prefix, slash := getPrefixAndSlashFromDaemonPlatform()

	tempDir, err := ioutil.TempDir("", "test-rm-container-with-removed-volume-")
	if err != nil {
		c.Fatalf("failed to create temporary directory: %s", tempDir)
	}
	defer os.RemoveAll(tempDir)

	dockerCmd(c, "run", "--name", "losemyvolumes", "-v", tempDir+":"+prefix+slash+"test", "busybox", "true")

	err = os.RemoveAll(tempDir)
	c.Assert(err, check.IsNil)

	dockerCmd(c, "rm", "-v", "losemyvolumes")
}

func (s *DockerSuite) TestRmContainerWithVolume(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	deleteAllContainers()
	prefix, slash := getPrefixAndSlashFromDaemonPlatform()

	dockerCmd(c, "run", "--name", "foo", "-v", prefix+slash+"srv", "busybox", "true")

	dockerCmd(c, "rm", "-v", "foo")
}

func (s *DockerSuite) TestRmContainerRunning(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	deleteAllContainers()
	createRunningContainer(c, "foo")

	time.Sleep(2 * time.Second)
	_, _, err := dockerCmdWithError("rm", "foo")
	c.Assert(err, checker.NotNil, check.Commentf("Expected error, can't rm a running container"))
}

func (s *DockerSuite) TestRmContainerForceRemoveRunning(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	deleteAllContainers()
	createRunningContainer(c, "foo")

	// Stop then remove with -s
	dockerCmd(c, "rm", "-f", "foo")
}

func (s *DockerSuite) TestRmInvalidContainer(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	out, _, err := dockerCmdWithError("rm", "unknown")
	c.Assert(err, checker.NotNil, check.Commentf("Expected error on rm unknown container, got none"))
	c.Assert(out, checker.Contains, "No such container")
}

func createRunningContainer(c *check.C, name string) {
	runSleepingContainer(c, "-dt", "--name", name)
	time.Sleep(1 * time.Second)
}
