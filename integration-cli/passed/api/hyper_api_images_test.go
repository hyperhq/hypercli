package main

import (
	"net/http"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestApiImagesSearchJSONContentType(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	testRequires(c, Network)

	res, b, err := sockRequestRaw("GET", "/images/search?term=test", nil, "application/json")
	c.Assert(err, check.IsNil)
	b.Close()
	c.Assert(res.StatusCode, checker.Equals, http.StatusOK)
	c.Assert(res.Header.Get("Content-Type"), checker.Equals, "application/json")
}
