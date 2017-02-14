package client

import (
	"fmt"
	// "io/ioutil"
	// "strconv"
	"strings"
	"time"
	"text/tabwriter"

	"github.com/docker/go-units"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/strslice"
	Cli "github.com/hyperhq/hypercli/cli"
	ropts "github.com/hyperhq/hypercli/opts"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	// "github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/runconfig/opts"
	"golang.org/x/net/context"
)

// CmdFunc is the parent subcommand for all func commands
//
// Usage: docker func <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdFunc(args ...string) error {
	cmd := Cli.Subcmd("func", []string{"COMMAND [OPTIONS]"}, funcUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

func funcUsage() string {
	funcCommands := [][]string{
		{"create", "Create a func"},
		{"update", "Update a func"},
		{"ls", "List all funcs"},
		{"rm", "Remove one or more funcs"},
		{"inspect", "Display detailed information on the given func"},
		{"call", "Call a func"},
		{"log", "Display execution log of a func"},
		{"get", "Query the request status of a func call"},
	}

	help := "Commands:\n"

	for _, cmd := range funcCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	help += fmt.Sprintf("\nRun 'hyper func COMMAND --help' for more information on a command.")
	return help
}

// CmdFuncCreate creates a new func with a given name
//
// Usage: hyper func create [OPTIONS] IMAGE [COMMAND]
func (cli *DockerCli) CmdFuncCreate(args ...string) error {
	cmd := Cli.Subcmd("func create", []string{"IMAGE [COMMAND]"}, "Create a new func", false)
	var (
		flName                = cmd.String([]string{"-name"}, "", "Func name")
		flSize                = cmd.String([]string{"-size"}, "s4", "The size of func containers (e.g. s1, s2, s3, s4, m1, m2, m3, l1, l2, l3)")
		flEnv                 = ropts.NewListOpts(opts.ValidateEnv)
		flHeader              = ropts.NewListOpts(opts.ValidateEnv)
		flMaxConcurrency      = cmd.Int([]string{"-max_concurrency"}, -1, "The maximum number of concurrent container, default (-1) is container quota")
		flMaxLimit            = cmd.Int([]string{"-max_limit"}, -1, "The maximum number of func call which waiting for completed, default (-1) is unlimit")
	)
	cmd.Var(&flEnv, []string{"e", "-env"}, "Set environment variables of container")
	cmd.Var(&flHeader, []string{"h", "-header"}, "The http response header of the endpoint of func status query")

	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	var (
		parsedArgs = cmd.Args()
		image      = cmd.Arg(0)
		command    strslice.StrSlice
	)

	if len(parsedArgs) > 1 {
		command = strslice.StrSlice(parsedArgs[1:])
	}

	// collect all the environment variables
	envVariables := flEnv.GetAll()

	// collect all the headers
	envHeaders := flHeader.GetAll()

	fnOpts := types.Func{
		Name:                *flName,
		Size:                *flSize,
		Image:               image,
		Command:             command,
		Env:                 &envVariables,
		Header:              &envHeaders,
		MaxConcurrency:      *flMaxConcurrency,
		MaxLimit:            *flMaxLimit,
	}

	fn, err := cli.client.FuncCreate(context.Background(), fnOpts)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "Func %s is created.\n", fn.Name)
	return nil
}

// CmdFuncDelete deletes one or more funcs
//
// Usage: hyper func rm NAME [NAME...]
func (cli *DockerCli) CmdFuncRm(args ...string) error {
	cmd := Cli.Subcmd("func rm", []string{"NAME [NAME...]"}, "Remove one or more funcs", false)
	cmd.Require(flag.Min, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	status := 0
	for _, fn := range cmd.Args() {
		if err := cli.client.FuncDelete(context.Background(), fn); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
		fmt.Fprintf(cli.out, "%s\n", fn)
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

// CmdFuncLs lists all the funcs
//
// Usage: hyper func ls [OPTIONS]
func (cli *DockerCli) CmdFuncLs(args ...string) error {
	cmd := Cli.Subcmd("func ls", nil, "Lists funcs", true)

	flFilter := ropts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process after get response from server.
	funcFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		if funcFilterArgs, err = filters.ParseFlag(f, funcFilterArgs); err != nil {
			return err
		}
	}

	options := types.FuncListOptions{
		Filters: funcFilterArgs,
	}

	fns, err := cli.client.FuncList(context.Background(), options)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tSIZE\tIMAGE\tCOMMAND\tCREATED\tUUID\n")
	for _, fn := range fns {
		created := units.HumanDuration(time.Now().UTC().Sub(fn.Created)) + " ago"
		command := strings.Join([]string(fn.Command), " ")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", fn.Name, fn.Size, fn.Image, command, created, fn.UUID)
	}

	w.Flush()
	return nil
}

// CmdFuncInspect
//
// Usage: docker func inspect [OPTIONS] NAME [NAME...]
func (cli *DockerCli) CmdFuncInspect(args ...string) error {
	cmd := Cli.Subcmd("func inspect", []string{"NAME [NAME...]"}, "Display detailed information on the given func", true)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	ctx := context.Background()

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		i, err := cli.client.FuncInspect(ctx, name)
		return i, nil, err
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}
