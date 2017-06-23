package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestCliLoadFromLocalDocker(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)

	testImage := "hello-world:latest"

	//local docker pull image
	pullCmd := exec.Command("docker", "-H", os.Getenv("LOCAL_DOCKER_HOST"), "pull", testImage)
	output, exitCode, err := runCommandWithOutput(pullCmd)
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//load image from local docker to hyper
	output, exitCode, err = dockerCmdWithError("load", "-l", testImage)
	c.Assert(output, checker.Contains, "has been loaded.")
	c.Assert(err, checker.IsNil)
	c.Assert(exitCode, checker.Equals, 0)

	//check image
	images, _ := dockerCmd(c, "images", "hello-world")
	c.Assert(images, checker.Contains, "hello-world")
}
