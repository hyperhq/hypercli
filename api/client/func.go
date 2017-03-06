package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-units"
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
		{"create", "Create a function"},
		{"update", "Update a function"},
		{"ls", "List all functions"},
		{"rm", "Remove one or more functions"},
		{"inspect", "Display detailed information on the given function"},
		{"call", "Call a function"},
		{"get", "Get the return of a function call"},
		{"logs", "Retrieve the logs of a function"},
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
	cmd := Cli.Subcmd("func create", []string{"IMAGE [COMMAND] [ARG...]"}, "Create a function", false)
	var (
		flName    = cmd.String([]string{"-name"}, "", "Function name")
		flSize    = cmd.String([]string{"-size"}, "s4", "The size of function containers to run the funciton (e.g. s1, s2, s3, s4, m1, m2, m3, l1, l2, l3)")
		flEnv     = ropts.NewListOpts(opts.ValidateEnv)
		flTimeout = cmd.Int([]string{"-timeout"}, 300, "The maximum execution duration of function call")
	)
	cmd.Var(&flEnv, []string{"e", "-env"}, "Set environment variables")

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

	fnOpts := types.Func{
		Name:    *flName,
		Size:    *flSize,
		Image:   image,
		Command: command,
		Env:     &envVariables,
		Timeout: *flTimeout,
	}

	fn, err := cli.client.FuncCreate(context.Background(), fnOpts)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s is created with the address of https://us-west-1.hyperfunc.io/%s/%s\n", fn.Name, fn.Name, fn.UUID)
	return nil
}

// CmdFuncUpdate updates a func with a given name
//
// Usage: hyper func update [OPTIONS] NAME
func (cli *DockerCli) CmdFuncUpdate(args ...string) error {
	cmd := Cli.Subcmd("func update", []string{"NAME"}, "Update a function", false)
	var (
		flSize    = cmd.String([]string{"-size"}, "", "The size of function containers to run the funciton (e.g. s1, s2, s3, s4, m1, m2, m3, l1, l2, l3)")
		flEnv     = ropts.NewListOpts(opts.ValidateEnv)
		flRefresh = cmd.Bool([]string{"-refresh"}, false, "Whether to regenerate the uuid of function")
		flTimeout = cmd.String([]string{"-timeout"}, "", "The maximum execution duration of function call")
	)
	cmd.Var(&flEnv, []string{"e", "-env"}, "Set environment variables")

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	name := cmd.Arg(0)

	// collect all the environment variables
	envVariables := flEnv.GetAll()

	timeout, _ := strconv.Atoi(*flTimeout)

	fnOpts := types.Func{
		Name:    name,
		Size:    *flSize,
		Env:     &envVariables,
		Refresh: *flRefresh,
		Timeout: timeout,
	}

	fn, err := cli.client.FuncUpdate(context.Background(), name, fnOpts)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "Function %s is updated.\n", fn.Name)
	return nil
}

// CmdFuncDelete deletes one or more funcs
//
// Usage: hyper func rm NAME [NAME...]
func (cli *DockerCli) CmdFuncRm(args ...string) error {
	cmd := Cli.Subcmd("func rm", []string{"NAME [NAME...]"}, "Remove one or more functions", false)
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
	cmd := Cli.Subcmd("func ls", nil, "Lists functions", true)

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
	cmd := Cli.Subcmd("func inspect", []string{"NAME [NAME...]"}, "Display detailed information on the given function", true)
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

// CmdFuncCall call a func
//
// Usage: hyper func call NAME
func (cli *DockerCli) CmdFuncCall(args ...string) error {
	cmd := Cli.Subcmd("func call", []string{"NAME"}, "Call a function", false)
	cmd.Require(flag.Exact, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	name := cmd.Arg(0)

	var stdin io.Reader
	if fi, err := os.Stdin.Stat(); err == nil {
		if fi.Mode()&os.ModeNamedPipe != 0 {
			stdin = bufio.NewReader(os.Stdin)
		}
	}

	ret, err := cli.client.FuncCall(context.Background(), name, stdin)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "CallId: %s\n", ret.CallId)
	return nil
}

// CmdFuncGet Get the return of a func call
//
// Usage: hyper func get [OPTIONS] CALL_ID
func (cli *DockerCli) CmdFuncGet(args ...string) error {
	cmd := Cli.Subcmd("func get", []string{"CALL_ID"}, "Get the return of a function call", false)
	wait := cmd.Bool([]string{"-wait"}, false, "Block until the call is completed")

	cmd.Require(flag.Exact, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	callId := cmd.Arg(0)

	ret, err := cli.client.FuncGet(context.Background(), callId, *wait)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s", ret)
	return nil
}

// CmdFuncLogs Get the return of a func call
//
// Usage: hyper func get [OPTIONS] NAME
func (cli *DockerCli) CmdFuncLogs(args ...string) error {
	cmd := Cli.Subcmd("func logs", []string{"NAME"}, "Retrieve the logs of a function", false)

	follow := cmd.Bool([]string{"f", "-follow"}, false, "Follow log output")
	tail := cmd.String([]string{"-tail"}, "all", "Number of lines to show from the end of the logs")
	callId := cmd.String([]string{"-callid"}, "", "Only retrieve specific logs of CallId")

	cmd.Require(flag.Exact, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	name := cmd.Arg(0)

	reader, err := cli.client.FuncLogs(context.Background(), name, *callId, *follow, *tail)
	if err != nil {
		return err
	}
	defer reader.Close()
	dec := json.NewDecoder(reader)
	for {
		var log types.FuncLogsResponse
		err := dec.Decode(&log)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if log.Event != "" {
			if log.Event == "CALL" {
				fmt.Fprintf(
					cli.out, "%s [%s] CallId: %s, ShortStdin: %s\n",
					log.Time, log.Event, log.CallId, log.ShortStdin,
				)
			} else if log.Event == "COMPLETED" {
				fmt.Fprintf(
					cli.out, "%s [%s] CallId: %s, ShortStdout: %s, ShortStderr: %s\n",
					log.Time, log.Event, log.CallId, log.ShortStdout, log.ShortStderr,
				)
			} else if log.Event == "COMPLETED" {
				fmt.Fprintf(
					cli.out, "%s [%s] CallId: %s, ShortStdout: %s, ShortStderr: %s\n",
					log.Time, log.Event, log.CallId, log.ShortStdout, log.ShortStderr,
				)
			} else if log.Event == "ERROR" {
				fmt.Fprintf(
					cli.out, "%s [%s] CallId: %s, Message: %s\n",
					log.Time, log.Event, log.CallId, log.Message,
				)
			} else {
				fmt.Fprintf(
					cli.out, "%s [%s] CallId: %s\n",
					log.Time, log.Event, log.CallId,
				)
			}
		}
	}
	return nil
}
