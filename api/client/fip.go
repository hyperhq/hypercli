package client

import (
	"fmt"
	"text/tabwriter"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/opts"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

// CmdFip is the parent subcommand for all fip commands
//
// Usage: docker fip <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdFip(args ...string) error {
	cmd := Cli.Subcmd("fip", []string{"COMMAND [OPTIONS]"}, fipUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdNetworkCreate creates a new fip with a given name
//
// Usage: docker fip create [OPTIONS] COUNT
func (cli *DockerCli) CmdFipAllocate(args ...string) error {
	cmd := Cli.Subcmd("fip allocate", []string{"COUNT"}, "Creates some new floating IPs by the user", false)

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	fips, err := cli.client.FipAllocate(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}
	for _, ip := range fips {
		fmt.Fprintf(cli.out, "%s\n", ip)
	}
	return nil
}

// CmdFipRelease deletes one or more fips
//
// Usage: docker fip release FIP [FIP...]
func (cli *DockerCli) CmdFipRelease(args ...string) error {
	cmd := Cli.Subcmd("fip release", []string{"FIP [FIP...]"}, "Release one or more fips", false)
	cmd.Require(flag.Min, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	status := 0
	for _, ip := range cmd.Args() {
		if err := cli.client.FipRelease(context.Background(), ip); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

// CmdFipAssociate connects a container to a floating IP
//
// Usage: docker fip associate [OPTIONS] <FIP> <CONTAINER>
func (cli *DockerCli) CmdFipAssociate(args ...string) error {
	cmd := Cli.Subcmd("fip associate", []string{"FIP CONTAINER"}, "Connects a container to a floating IP", false)
	cmd.Require(flag.Min, 2)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}
	return cli.client.FipAssociate(context.Background(), cmd.Arg(0), cmd.Arg(1))
}

// CmdFipDisassociate disconnects a container from a floating IP
//
// Usage: docker fip disassociate <CONTAINER>
func (cli *DockerCli) CmdFipDisassociate(args ...string) error {
	cmd := Cli.Subcmd("fip disassociate", []string{"CONTAINER"}, "Disconnects container from a floating IP", false)
	//force := cmd.Bool([]string{"f", "-force"}, false, "Force the container to disconnect from a floating IP")
	cmd.Require(flag.Exact, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	ip, err := cli.client.FipDisassociate(context.Background(), cmd.Arg(0))
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s\n", ip)
	return nil
}

// CmdFipLs lists all the fips
//
// Usage: docker fip ls [OPTIONS]
func (cli *DockerCli) CmdFipLs(args ...string) error {
	cmd := Cli.Subcmd("fip ls", nil, "Lists fips", true)

	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process after get response from server.
	fipFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		if fipFilterArgs, err = filters.ParseFlag(f, fipFilterArgs); err != nil {
			return err
		}
	}

	options := types.NetworkListOptions{
		Filters: fipFilterArgs,
	}

	fips, err := cli.client.FipList(context.Background(), options)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "Floating IP\tContainer")
	fmt.Fprintf(w, "\n")
	for _, fip := range fips {
		fmt.Fprintf(w, "%s\t%s\n", fip["fip"], fip["container"])
	}

	w.Flush()
	return nil
}

func fipUsage() string {
	fipCommands := [][]string{
		{"allocate", "Allocate a or some IPs"},
		{"associate", "Associate floating IP to container"},
		{"disassociate", "Disassociate floating IP from conainer"},
		{"ls", "List all floating IPs"},
		{"release", "Release a floating IP"},
	}

	help := "Commands:\n"

	for _, cmd := range fipCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	help += fmt.Sprintf("\nRun 'hyper fip COMMAND --help' for more information on a command.")
	return help
}
