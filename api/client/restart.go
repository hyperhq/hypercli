package client

import (
	"fmt"
	"strings"

	Cli "github.com/hyperhq/hypercli/cli"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

func (cli *DockerCli) restartReloadContainer(name string, seconds int) error {
	// stop container
	err := cli.client.ContainerStop(name, seconds)
	if err != nil {
		return err
	}
	// reload all sourced volumes of a container
	initvols, err := cli.containerFilterInitVolumes(name)
	if err != nil {
		return err
	}
	if len(initvols) > 0 {
		err = cli.containerReloadInitVolumes(initvols)
		if err != nil {
			return err
		}
	}
	// start container
	return cli.client.ContainerStart(name)
}

// CmdRestart restarts one or more containers.
//
// Usage: docker restart [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdRestart(args ...string) error {
	cmd := Cli.Subcmd("restart", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["restart"].Description, true)
	nSeconds := cmd.Int([]string{"t", "-time"}, 10, "Seconds to wait for stop before killing the container")
	reload := cmd.Bool([]string{"r", "-reload"}, false, "Reload container's volumes that have a source")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	if *reload {
		addTrustedFlags(cmd, true)
	}
	var errs []string
	for _, name := range cmd.Args() {
		restartFunc := cli.client.ContainerRestart
		if *reload {
			restartFunc = cli.restartReloadContainer
		}
		if err := restartFunc(name, *nSeconds); err != nil {
			errs = append(errs, err.Error())
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
