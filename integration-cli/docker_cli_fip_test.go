package main

import (
	"string"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestFipLs(c *check.C) {
	out, _ := runSleepingContainer(c, "-d")
	firstID := strings.TrimSpace(out)

}
