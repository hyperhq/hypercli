package main

import (
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestShareNamedVolume(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	volName := "testvolume"
	_, err := dockerCmd(c, "run", "-d", "--name=volserver", "-v", volName+":/data", "hyperhq/nfs-server")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "exec", "volclient", "ls", "/data")
	c.Assert(err, checker.Equals, 0)
}

func (s *DockerSuite) TestShareImplicitVolume(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	_, err := dockerCmd(c, "run", "-d", "--name=volserver", "-v", "/data", "hyperhq/nfs-server")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "exec", "volclient", "ls", "/data")
	c.Assert(err, checker.Equals, 0)
}

func (s *DockerSuite) TestSharePopulatedVolume(c *check.C) {
        printTestCaseName()
        defer printTestDuration(time.Now())
        _, err := dockerCmd(c, "run", "-d", "--name=volserver", "-v", "https://github.com/hyperhq/hypercli.git:/data", "hyperhq/nfs-server")
        c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
	c.Assert(err, checker.Equals, 0)
        out, err := dockerCmd(c, "exec", "volclient", "ls", "/data")
        c.Assert(err, checker.Equals, 0)
        c.Assert(out, checker.Contains, "Dockerfile")
}

func (s *DockerSuite) TestShareVolumeBadSource(c *check.C) {
	printTestCaseName()
        defer printTestDuration(time.Now())
        _, err := dockerCmd(c, "run", "-d", "--name=volserver", "-v", "/data", "busybox")
        c.Assert(err, checker.Equals, 0)
	_, _, failErr := dockerCmdWithError("run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
        c.Assert(failErr, checker.NotNil)
}

func (s *DockerSuite) TestShareVolumeNoSource(c *check.C) {
	printTestCaseName()
        defer printTestDuration(time.Now())
	_, _, err := dockerCmdWithError("run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
        c.Assert(err, checker.NotNil)
}

func (s *DockerSuite) TestShareNoVolume(c *check.C) {
        printTestCaseName()
        defer printTestDuration(time.Now())
        _, err := dockerCmd(c, "run", "-d", "--name=volserver", "hyperhq/nfs-server")
        c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "run", "-d", "--name=volclient", "--volumes-from", "volserver", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, _, failErr := dockerCmdWithError("exec", "volclient", "ls", "/data")
        c.Assert(failErr, checker.NotNil)
}
