package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

const (
	singlePortImage      = "hyperhq/test-port-single"
	multiPortImage       = "hyperhq/test-port-list"
	mixPortocalPortImage = "hyperhq/test-port-mix"
)

func (s *DockerSuite) TestPortList(c *check.C) {
	testRequires(c, DaemonIsLinux)

	// one port
	//	_, errCode := dockerCmd(c, "pull", singlePortImage)
	//	c.Assert(errCode, checker.Equals, 0)

	out, _ := dockerCmd(c, "run", "-d", singlePortImage, "top")
	firstID := strings.TrimSpace(out)

	out, _ = dockerCmd(c, "port", firstID, "80")

	err := assertPortList(c, out, []string{"0.0.0.0:80"})
	// Port list is not correct
	c.Assert(err, checker.IsNil)

	out, _ = dockerCmd(c, "port", firstID)

	err = assertPortList(c, out, []string{"80/tcp -> 0.0.0.0:80"})
	// Port list is not correct
	c.Assert(err, checker.IsNil)

	dockerCmd(c, "rm", "-f", firstID)

	// four port
	out, _ = dockerCmd(c, "run", "-d",
		multiPortImage, "top")
	ID := strings.TrimSpace(out)

	out, _ = dockerCmd(c, "port", ID, "80")

	err = assertPortList(c, out, []string{"0.0.0.0:80"})
	// Port list is not correct
	c.Assert(err, checker.IsNil)

	out, _ = dockerCmd(c, "port", ID)

	err = assertPortList(c, out, []string{
		"80/tcp -> 0.0.0.0:80",
		"82/tcp -> 0.0.0.0:82",
		"84/tcp -> 0.0.0.0:84",
		"86/tcp -> 0.0.0.0:86"})
	// Port list is not correct
	c.Assert(err, checker.IsNil)

	dockerCmd(c, "rm", "-f", ID)

	// test mixing protocols in same port range
	out, _ = dockerCmd(c, "run", "-d",
		mixPortocalPortImage, "top")
	ID = strings.TrimSpace(out)

	out, _ = dockerCmd(c, "port", ID)

	err = assertPortList(c, out, []string{
		"80/tcp -> 0.0.0.0:80",
		"81/udp -> 0.0.0.0:81"})
	// Port list is not correct
	c.Assert(err, checker.IsNil)
	dockerCmd(c, "rm", "-f", ID)
}

func assertPortList(c *check.C, out string, expected []string) error {
	lines := strings.Split(strings.Trim(out, "\n "), "\n")
	if len(lines) != len(expected) {
		return fmt.Errorf("different size lists %s, %d, %d", out, len(lines), len(expected))
	}
	sort.Strings(lines)
	sort.Strings(expected)

	for i := 0; i < len(expected); i++ {
		if lines[i] != expected[i] {
			return fmt.Errorf("|" + lines[i] + "!=" + expected[i] + "|")
		}
	}

	return nil
}

func (s *DockerSuite) TestUnpublishedPortsInPsOutput(c *check.C) {
	testRequires(c, DaemonIsLinux)
	port1 := 80
	port2 := 82
	unpPort1 := fmt.Sprintf("%d/tcp", port1)
	unpPort2 := fmt.Sprintf("%d/tcp", port2)

	// Run the container auto publish the exposed ports
	dockerCmd(c, "run", "-d", multiPortImage, "sleep", "35")

	// Check docker ps o/p for last created container reports the exposed ports in the port bindings
	expBnd1 := fmt.Sprintf("0.0.0.0:%d->%s", port1, unpPort1)
	expBnd2 := fmt.Sprintf("0.0.0.0:%d->%s", port2, unpPort2)
	out, _ := dockerCmd(c, "ps", "-n=1")
	// Cannot find expected port binding port (0.0.0.0:xxxxx->unpPort1) in docker ps output
	c.Assert(strings.Contains(out, expBnd1), checker.Equals, true, check.Commentf("out: %s; unpPort1: %s", out, unpPort1))
	// Cannot find expected port binding port (0.0.0.0:xxxxx->unpPort2) in docker ps output
	c.Assert(strings.Contains(out, expBnd2), checker.Equals, true, check.Commentf("out: %s; unpPort2: %s", out, unpPort2))
}
