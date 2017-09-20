package main

import (
	"time"
	"os"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestApiContainerStartNilHostconfig(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())
	testRequires(c, DaemonIsLinux)
	name := "testing"
	config := map[string]interface{}{
		"Image": "busybox",
	}

	_, _, err := sockRequest("POST", "/containers/create?name="+name, config, os.Getenv("REGION"))
	c.Assert(err, checker.IsNil)

	config = map[string]interface{}{}
	_, _, err = sockRequest("POST", "/containers/"+name+"/start", config, os.Getenv("REGION"))
	c.Assert(err, checker.IsNil)
}
