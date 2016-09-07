package client

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/opts"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

// CmdUpdate updates resources of one or more containers.
//
// Usage: hyper update [OPTIONS] CONTAINER [CONTAINER...]
func (cli *DockerCli) CmdUpdate(args ...string) error {
	cmd := Cli.Subcmd("update", []string{"CONTAINER [CONTAINER...]"}, Cli.DockerCommands["update"].Description, true)
	flSecurityGroups := opts.NewListOpts(nil)
	cmd.Var(&flSecurityGroups, []string{"-sg"}, "Security group for each container")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)
	if cmd.NFlag() == 0 {
		return fmt.Errorf("You must provide one or more flags when using this command.")
	}

	ctx := context.Background()
	names := cmd.Args()
	var errs []string
	for _, name := range names {
		res, err := cli.client.ContainerInspect(ctx, name)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		labels := map[string]string{}
		for label, val := range res.Config.Labels {
			if !strings.HasPrefix(label, "sh_hyper_sg_") {
				labels[label] = val
			}
		}
		for _, label := range flSecurityGroups.GetAll() {
			if label == "" {
				continue
			}
			labels["sh_hyper_sg_"+label] = "yes"
		}
		var updateConfig struct {
			Labels map[string]string
		}
		updateConfig.Labels = labels
		if err := cli.client.ContainerUpdate(ctx, name, updateConfig); err != nil {
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
