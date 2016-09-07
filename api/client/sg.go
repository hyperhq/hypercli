package client

import (
	"fmt"
	"os"
	"text/tabwriter"

	Cli "github.com/hyperhq/hypercli/cli"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

// CmdSg is the parent subcommand for all sg commands
//
// Usage: docker sg <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdSg(args ...string) error {
	cmd := Cli.Subcmd("sg", []string{"COMMAND [OPTIONS]"}, sgUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdSgCreate creates a new sg with a given name
//
// Usage: hyper sg create [OPTIONS] NAME
func (cli *DockerCli) CmdSgCreate(args ...string) error {
	cmd := Cli.Subcmd("sg create", []string{"NAME"}, "Creates some new security groups by the user", false)
	file := cmd.String([]string{"f", "-file"}, "", "Yaml file to create security group")

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	data, err := os.Open(*file)
	if err != nil {
		return err
	}

	err = cli.client.SgCreate(context.Background(), cmd.Arg(0), data)
	if err != nil {
		return err
	}
	return nil
}

// CmdSgRm removes a sg with a given name
//
// Usage: hyper sg rm [OPTIONS] NAME
func (cli *DockerCli) CmdSgRm(args ...string) error {
	cmd := Cli.Subcmd("sg rm", []string{"NAME"}, "Removes some security groups by the user", false)

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	err = cli.client.SgRm(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}
	return nil
}

// CmdSgLs list security groups
//
// Usage: hyper sg ls [OPTIONS]
func (cli *DockerCli) CmdSgLs(args ...string) error {
	cmd := Cli.Subcmd("sg ls", []string{}, "List security groups by the user", false)

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	sgs, err := cli.client.SgLs(context.Background())
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "Name\tDescription")
	fmt.Fprintf(w, "\n")
	for _, sg := range sgs {
		fmt.Fprintf(w, "%s\t%s\n", sg.GroupName, sg.Description)
	}
	w.Flush()
	return nil
}

// CmdSgInspect Inspect security groups
//
// Usage: hyper sg inspect [OPTIONS] NAME
func (cli *DockerCli) CmdSgInspect(args ...string) error {
	cmd := Cli.Subcmd("sg inspect", []string{"NAME"}, "Inspect the security group by the user", false)

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	sg, err := cli.client.SgInspect(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}
	output, err := yaml.Marshal(sg)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(output))
	return nil
}

// CmdSgUpdate Update the security group
//
// Usage: hyper sg update [OPTIONS] NAME
func (cli *DockerCli) CmdSgUpdate(args ...string) error {
	cmd := Cli.Subcmd("sg update", []string{"NAME"}, "Update the security group by the user", false)
	file := cmd.String([]string{"f", "-file"}, "", "Yaml file to update security group")

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	data, err := os.Open(*file)
	if err != nil {
		return err
	}

	err = cli.client.SgUpdate(context.Background(), cmd.Arg(0), data)
	if err != nil {
		return err
	}
	return nil
}

func sgUsage() string {
	sgCommands := [][]string{
		{"create", "Create security group"},
		{"ls", "List all security groups"},
		{"rm", "Remove security group"},
		{"inspect", "Inspect the security group"},
		{"update", "Update the security group config"},
	}

	help := "Commands:\n"

	for _, cmd := range sgCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	help += fmt.Sprintf("\nRun 'hyper sg COMMAND --help' for more information on a command.")
	return help
}
