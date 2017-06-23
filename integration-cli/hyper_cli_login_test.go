package main

import (
	"bytes"
	"os"
	"os/exec"
	"time"
	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestCliLoginWithoutTTY(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	cmd := exec.Command(dockerBinary, "-H", os.Getenv("DOCKER_HOST"), "login")

	// Send to stdin so the process does not get the TTY
	cmd.Stdin = bytes.NewBufferString("buffer test string \n")

	// run the command and block until it's done
	err := cmd.Run()
	c.Assert(err, checker.NotNil) //"Expected non nil err when loginning in & TTY not available"
}

/*
// Hyper can not login to private registry
func (s *DockerRegistryAuthSuite) TestCliLoginToPrivateRegistry(c *check.C) {
	printTestCaseName(); defer printTestDuration(time.Now())
	// wrong credentials
	out, _, err := dockerCmdWithError("login", "-u", s.reg.username, "-p", "WRONGPASSWORD", "-e", s.reg.email, privateRegistryURL)
	c.Assert(err, checker.NotNil, check.Commentf(out))
	c.Assert(out, checker.Contains, "401 Unauthorized")

	// now it's fine
	dockerCmd(c, "login", "-u", s.reg.username, "-p", s.reg.password, "-e", s.reg.email, privateRegistryURL)
}
*/
