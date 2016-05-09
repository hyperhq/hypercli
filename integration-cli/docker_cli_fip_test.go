package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

//TODO: release a non-IP
func (s *DockerSuite) TestAssociateUsedIP(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fmt.Println(firstIP)
	fipList := []string{firstIP}
	defer releaseFip(c, fipList)

	out, _ = runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)
	fmt.Println(firstContainerID)

	out, _ = runSleepingContainer(c, "-d")
	secondContainerID := strings.TrimSpace(out)
	fmt.Println(secondContainerID)

	dockerCmd(c, "fip", "associate", firstIP, firstContainerID)
	out, _, err := dockerCmdWithError("fip", "associate", firstIP, secondContainerID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstContainerID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}

func (s *DockerSuite) TestAssociateConfedContainer(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fmt.Println(firstIP)
	fipList := []string{firstIP}

	out, _ = dockerCmd(c, "fip", "allocate", "1")
	secondIP := strings.TrimSpace(out)
	fipList = append(fipList, secondIP)
	defer releaseFip(c, fipList)

	out, _ = runSleepingContainer(c, "-d")
	firstContainerID := strings.TrimSpace(out)
	fmt.Println(firstContainerID)

	dockerCmd(c, "fip", "associate", firstIP, firstContainerID)
	out, _, err := dockerCmdWithError("fip", "associate", secondIP, firstContainerID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstContainerID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}
