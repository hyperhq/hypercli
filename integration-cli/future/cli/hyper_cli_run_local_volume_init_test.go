package main

import (
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestRunLocalFileVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	source := "/tmp/hyper_integration_test_local_file_volume_file"
	ioutil.WriteFile(source, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", source+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data")
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")
	dockerCmd(c, "rm", "-fv", "voltest")

	exec.Command("rm", "-f", source).CombinedOutput()
}

func (s *DockerSuite) TestRunLocalDirVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	dir := "/tmp/hyper_integration_test_local_dir_volume_dir"
	file := "datafile"
	exec.Command("mkdir", "-p", dir).CombinedOutput()
	ioutil.WriteFile(dir+"/"+file, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", dir+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/"+file)
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")
	dockerCmd(c, "rm", "-fv", "voltest")

	exec.Command("rm", "-r", dir).CombinedOutput()
}

func (s *DockerSuite) TestRunLocalDeepDirVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	dir := "/tmp/hyper_integration_test_local_dir_volume_dir"
	middle_dir := "/dir1/dir2/dir3/dir4/dir5"
	file := "datafile"
	exec.Command("mkdir", "-p", dir+"/"+middle_dir).CombinedOutput()
	ioutil.WriteFile(dir+"/"+middle_dir+"/"+file, []byte("foo"), 0644)

	_, err := dockerCmd(c, "run", "-d", "--name=voltest", "-v", dir+":/data", "busybox")
	c.Assert(err, checker.Equals, 0)
	out, err := dockerCmd(c, "exec", "voltest", "cat", "/data/"+middle_dir+"/"+file)
	c.Assert(err, checker.Equals, 0)
	c.Assert(out, checker.Equals, "foo")
	dockerCmd(c, "rm", "-fv", "voltest")

	exec.Command("rm", "-r", dir).CombinedOutput()
}

func (s *DockerSuite) TestRunLocalNonexistingVolumeBinding(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	dir := "/tmp/nosuchfile"
	_, _, err := dockerCmdWithError("run", "-d", "--name=voltest", "-v", dir+":/data", "busybox")
	c.Assert(err, checker.NotNil)
}
