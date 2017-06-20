package main

import (
	"strings"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestCliFipAssociateUsedIP(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fipList := []string{firstIP}
	defer releaseFip(c, fipList)

	pullImageIfNotExist("busybox")
	out, _ = runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)

	out, _ = runSleepingContainer(c, "-d")
	secondContainerID := strings.TrimSpace(out)

	dockerCmd(c, "fip", "associate", firstIP, firstContainerID)
	out, _, err := dockerCmdWithError("fip", "associate", firstIP, secondContainerID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstContainerID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}

func (s *DockerSuite) TestCliFipAssociateConfedContainer(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fipList := []string{firstIP}

	out, _ = dockerCmd(c, "fip", "allocate", "1")
	secondIP := strings.TrimSpace(out)
	fipList = append(fipList, secondIP)
	defer releaseFip(c, fipList)

	pullImageIfNotExist("busybox")
	out, _ = runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)

	dockerCmd(c, "fip", "associate", firstIP, firstContainerID)
	out, _, err := dockerCmdWithError("fip", "associate", secondIP, firstContainerID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstContainerID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}

func (s *DockerSuite) TestCliFipDisassociateUnconfedContainer(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	pullImageIfNotExist("busybox")
	out, _ := runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)

	out, _, err := dockerCmdWithError("fip", "disassociate", firstContainerID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
}

func (s *DockerSuite) TestCliFipReleaseUsedIP(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fipList := []string{firstIP}
	defer releaseFip(c, fipList)

	pullImageIfNotExist("busybox")
	out, _ = runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)

	dockerCmd(c, "fip", "associate", firstIP, firstContainerID)
	out, _, err := dockerCmdWithError("fip", "release", firstIP)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstContainerID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}

func (s *DockerSuite) TestCliFipReleaseInvalidIP(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _, err := dockerCmdWithError("fip", "release", "InvalidIP")
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))

	out, _, err = dockerCmdWithError("fip", "release", "0.0.0.0")
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
}
