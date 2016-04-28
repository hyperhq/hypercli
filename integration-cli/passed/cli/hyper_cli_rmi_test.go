package main

import (
	"strings"
	"time"
	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestRmiWithContainerFails(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	errSubstr := "is using it"

	// create a container
	out, _ := dockerCmd(c, "run", "-d", "busybox", "true")

	cleanedContainerID := strings.TrimSpace(out)

	// try to delete the image
	out, _, err := dockerCmdWithError("rmi", "busybox")
	// Container is using image, should not be able to rmi
	c.Assert(err, checker.NotNil)
	// Container is using image, error message should contain errSubstr
	c.Assert(out, checker.Contains, errSubstr, check.Commentf("Container: %q", cleanedContainerID))

	// make sure it didn't delete the busybox name
	images, _ := dockerCmd(c, "images")
	// The name 'busybox' should not have been removed from images
	c.Assert(images, checker.Contains, "busybox")
}



func (s *DockerSuite) TestRmiBlank(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	// try to delete a blank image name
	out, _, err := dockerCmdWithError("rmi", "")
	// Should have failed to delete '' image
	c.Assert(err, checker.NotNil)
	// Wrong error message generated
	c.Assert(out, checker.Not(checker.Contains), "no such id", check.Commentf("out: %s", out))
	// Expected error message not generated
	c.Assert(out, checker.Contains, "image name cannot be blank\n", check.Commentf("out: %s", out))

	out, _, err = dockerCmdWithError("rmi", " ")
	// Should have failed to delete ' ' image
	c.Assert(err, checker.NotNil)
	// Expected error message not generated
	c.Assert(out, checker.Contains, "image name cannot be blank\n", check.Commentf("out: %s", out))
}



// #18873
func (s *DockerSuite) TestRmiByIDHardConflict(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	// TODO Windows CI. This will work on a TP5 compatible docker which
	// has content addressibility fixes. Do not run this on TP4 as it
	// will end up deleting the busybox image causing subsequent tests to fail.
	testRequires(c, DaemonIsLinux)
	dockerCmd(c, "create", "busybox")

	imgID := inspectField(c, "busybox:latest", "Id")

	_, _, err := dockerCmdWithError("rmi", imgID[:12])
	c.Assert(err, checker.NotNil)

	// check that tag was not removed
	imgID2 := inspectField(c, "busybox:latest", "Id")
	c.Assert(imgID, checker.Equals, imgID2)
}
