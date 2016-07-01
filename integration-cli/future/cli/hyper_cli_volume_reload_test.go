package main

import (
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestRunHttpFileVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://raw.githubusercontent.com/hyperhq/hypercli/master/README.md"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "stat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "regular file")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsFileVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://raw.githubusercontent.com/hyperhq/hypercli/master/README.md"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "stat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "regular file")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunGitVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "git://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunGitBranchVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "git://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/configure.ac")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "2.13.0")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpGitVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpGitBranchVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/configure.ac")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "2.13.0")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsGitVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "https://git.kernel.org/pub/scm/utils/util-linux/util-linux.git"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunHttpsGitBranchVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "http://git.kernel.org/pub/scm/utils/util-linux/util-linux.git:stable/v2.13.0"
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/README")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Contains, "util-linux")
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunNoVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "busybox")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	dockerCmd(c, "rm", "-fv", "voltest")
}

func (s *DockerSuite) TestRunLocalFileVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "/tmp/hyper_integration_test_local_file_volume_file"
	ioutil.WriteFile(source, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")

	ioutil.WriteFile(source, []byte("bar"), 0644)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err = dockerCmd(c, "exec", "voltest", "cat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "bar")

	dockerCmd(c, "rm", "-fv", "voltest")
	exec.Command("rm", "-f", source).Output()
}

func (s *DockerSuite) TestRunLocalDirVolumeRestartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	dir := "/tmp/hyper_integration_test_local_dir_volume_dir"
	file := "datafile"
	exec.Command("mkdir", "-p", dir).Output()
	ioutil.WriteFile(dir+"/"+file, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", dir+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/"+file)
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")

	ioutil.WriteFile(dir+"/"+file, []byte("bar"), 0644)
	_, err = dockerCmd(c, "restart", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err = dockerCmd(c, "exec", "voltest", "cat", "/data/"+file)
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "bar")

	dockerCmd(c, "rm", "-fv", "voltest")
	exec.Command("rm", "-r", dir).Output()
}

func (s *DockerSuite) TestRunLocalFileVolumeStartReload(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "/tmp/hyper_integration_test_local_file_volume_file"
	ioutil.WriteFile(source, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")

	ioutil.WriteFile(source, []byte("bar"), 0644)
	_, err = dockerCmd(c, "stop", "voltest")
	c.Assert(err, checker.Equals, 0)
	_, err = dockerCmd(c, "start", "--reload")
	c.Assert(err, checker.Equals, 0)
	out, err = dockerCmd(c, "exec", "voltest", "cat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "bar")

	dockerCmd(c, "rm", "-fv", "voltest")
	exec.Command("rm", "-f", source).Output()
}
