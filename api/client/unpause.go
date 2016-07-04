package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/hyperhq/hypercli/cli"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

// CmdUnpause unpauses all processes within a container, for one or more containers.
//
// Usage: docker unpause CONTAINER [CONTAINER...]
func (cli *DockerCli) Unpause(args ...string) error {
	cmd := Cli.Subcmd("unpause", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["unpause"].Description, true)
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	ctx := context.Background()

	var errs []string
	for _, name := range cmd.Args() {
		if err := cli.client.ContainerUnpause(ctx, name); err != nil {
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
