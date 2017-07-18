package main

import (
	"path/filepath"
	"time"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
	"github.com/hyperhq/hypercli/cliconfig"
	"github.com/hyperhq/hypercli/pkg/homedir"
)

func (s *DockerSuite) TestCliConfigAndRewrite(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "config", "--accesskey", "xx", "--secretkey", "xxxx", "tcp://127.0.0.1:6443")
	c.Assert(out, checker.Contains, "WARNING: Your login credentials has been saved in " + homedir.Get() + "/.hyper/config.json")

	configDir := filepath.Join(homedir.Get(), ".hyper")
	conf, err := cliconfig.Load(configDir)
	c.Assert(err, checker.IsNil)
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey, checker.Equals, "xx", check.Commentf("Should get xx, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey))
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey, checker.Equals, "xxxx", check.Commentf("Should get xxxx, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey))

	out, _ = dockerCmd(c, "config", "--accesskey", "yy", "--secretkey", "yyyy", "tcp://127.0.0.1:6443")
	c.Assert(out, checker.Contains, "WARNING: Your login credentials has been saved in " + homedir.Get() + "/.hyper/config.json")

	conf, err = cliconfig.Load(configDir)
	c.Assert(err, checker.IsNil)
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey, checker.Equals, "yy", check.Commentf("Should get yy, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey))
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey, checker.Equals, "yyyy", check.Commentf("Should get yyyy, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey))
}

func (s *DockerSuite) TestCliConfigMultiHost(c *check.C) {
	printTestCaseName()
	defer printTestDuration(time.Now())

	out, _ := dockerCmd(c, "config", "--accesskey", "xx", "--secretkey", "xxxx", "tcp://127.0.0.1:6443")
	c.Assert(out, checker.Contains, "WARNING: Your login credentials has been saved in " + homedir.Get() + "/.hyper/config.json")

	configDir := filepath.Join(homedir.Get(), ".hyper")
	conf, err := cliconfig.Load(configDir)
	c.Assert(err, checker.IsNil)
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey, checker.Equals, "xx", check.Commentf("Should get xx, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].AccessKey))
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey, checker.Equals, "xxxx", check.Commentf("Should get xxxx, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6443"].SecretKey))

	out, _ = dockerCmd(c, "config", "--accesskey", "yy", "--secretkey", "yyyy", "tcp://127.0.0.1:6444")
	c.Assert(out, checker.Contains, "WARNING: Your login credentials has been saved in " + homedir.Get() + "/.hyper/config.json")

	conf, err = cliconfig.Load(configDir)
	c.Assert(err, checker.IsNil)
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6444"].AccessKey, checker.Equals, "yy", check.Commentf("Should get yy, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6444"].AccessKey))
	c.Assert(conf.CloudConfig["tcp://127.0.0.1:6444"].SecretKey, checker.Equals, "yyyy", check.Commentf("Should get yyyy, but get %s\n", conf.CloudConfig["tcp://127.0.0.1:6444"].SecretKey))
}
