package main

import (
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestRunHttpFileVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://raw.githubusercontent.com/hyperhq/hypercli/master/README.md"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "stat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "regular file")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsFileVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://raw.githubusercontent.com/hyperhq/hypercli/master/README.md"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "stat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "regular file")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunGitVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "git://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunGitBranchVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "git://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/configure.ac")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "2.13.0")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpGitVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpGitBranchVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/configure.ac")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "2.13.0")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsGitVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsGitBranchVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunNonexistingHttpFileVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://raw.githubusercontent.com/nosuchuser/nosuchrepo/masterbeta/README.md"
	_, _, err := dockerCmdWithError("run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.NotNil)
}
