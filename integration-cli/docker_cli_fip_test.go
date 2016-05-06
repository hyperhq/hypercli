package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestAssociateUsedIP(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "fip", "allocate", "1")
	firstIP := strings.TrimSpace(out)
	fmt.Println(firstIP)
	fipList := []string{firstIP}
	defer releaseFip(c, fipList)

	out, _ = runSleepingContainer(c, "-d")
	firstID := strings.TrimSpace(out)
	fmt.Println(firstID)

	out, _ = runSleepingContainer(c, "-d")
	secondID := strings.TrimSpace(out)
	fmt.Println(secondID)

	dockerCmd(c, "fip", "associate", firstIP, firstID)
	out, _, err := dockerCmdWithError("fip", "associate", firstIP, secondID)
	c.Assert(err, checker.NotNil, check.Commentf("Should fail.", out, err))
	out, _ = dockerCmd(c, "fip", "disassociate", firstID)
	c.Assert(out, checker.Equals, firstIP+"\n")
}
