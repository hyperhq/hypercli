package client

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
	Cli "github.com/hyperhq/hypercli/cli"
	ropts "github.com/hyperhq/hypercli/opts"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	"github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/runconfig/opts"
	"golang.org/x/net/context"
)

// CmdCron is the parent subcommand for all cron commands
//
// Usage: docker cron <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdCron(args ...string) error {
	cmd := Cli.Subcmd("cron", []string{"COMMAND [OPTIONS]"}, cronUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdCronCreate creates a new cron with a given name
//
// Usage: hyper cron create [OPTIONS]
func (cli *DockerCli) CmdCronCreate(args ...string) error {
	cmd := Cli.Subcmd("cron create", []string{"IMAGE"}, "Create a new cron", false)
	var (
		flSecurityGroups = ropts.NewListOpts(nil)
		flEnv            = ropts.NewListOpts(opts.ValidateEnv)
		flLabels         = ropts.NewListOpts(opts.ValidateEnv)
		flEnvFile        = ropts.NewListOpts(nil)
		flVolumes        = ropts.NewListOpts(nil)
		flLinks          = ropts.NewListOpts(opts.ValidateLink)
		flLabelsFile     = ropts.NewListOpts(nil)
		flPublish        = ropts.NewListOpts(nil)
		flExpose         = ropts.NewListOpts(nil)

		flName          = cmd.String([]string{"-name"}, "", "Cron name")
		flContainerName = cmd.String([]string{"-container-name"}, "", "Cron container name")
		flEntrypoint    = cmd.String([]string{"-entrypoint"}, "", "Overwrite the default ENTRYPOINT of the image")
		flNetMode       = cmd.String([]string{}, "bridge", "Connect containers to a network, only bridge is supported now")
		flStopSignal    = cmd.String([]string{"-stop-signal"}, signal.DefaultStopSignal, fmt.Sprintf("Signal to stop a container, %v by default", signal.DefaultStopSignal))
		flContainerSize = cmd.String([]string{"-size"}, "s4", "The size of cron containers (e.g. s1, s2, s3, s4, m1, m2, m3, l1, l2, l3)")
		flWorkingDir    = cmd.String([]string{"w", "-workdir"}, "", "Working directory inside the container")
		flHostname      = cmd.String([]string{"h", "-hostname"}, "", "Container host name")
		flNoAutoVolume  = cmd.Bool([]string{"-noauto-volume"}, false, "Do not create volumes specified in image")
		flPublishAll    = cmd.Bool([]string{"P", "-publish-all"}, false, "Publish all exposed ports to random ports")
		flRestartPolicy = cmd.String([]string{"-restart"}, "no", "Restart policy to apply when a container exits")
		flMailTo        = cmd.String([]string{"-mailto"}, "", "Mail to while the cron has something")

		flMinute = cmd.String([]string{"-minute"}, "0", "minute")
		flHour   = cmd.String([]string{"-hour"}, "0", "hour")
		flDom    = cmd.String([]string{"-dom"}, "*", "dom")
		flDow    = cmd.String([]string{"-week"}, "*", "dow")
		flMonth  = cmd.String([]string{"-month"}, "*", "month")
	)
	cmd.Var(&flLabels, []string{"l", "-label"}, "Set meta data on a container")
	cmd.Var(&flLabelsFile, []string{"-label-file"}, "Read in a line delimited file of labels")
	cmd.Var(&flEnv, []string{"e", "-env"}, "Set environment variables")
	cmd.Var(&flEnvFile, []string{"-env-file"}, "Read in a file of environment variables")
	cmd.Var(&flSecurityGroups, []string{"-sg"}, "Security group for each container")
	cmd.Var(&flVolumes, []string{"v", "--volume"}, "Volume for each container")
	cmd.Var(&flLinks, []string{"-link"}, "Add link to another container")
	cmd.Var(&flPublish, []string{"p", "-publish"}, "Publish a container's port(s) to the host")
	cmd.Var(&flExpose, []string{"-expose"}, "Expose a port or a range of ports")

	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	var (
		parsedArgs = cmd.Args()
		runCmd     strslice.StrSlice
		entrypoint strslice.StrSlice
		image      = cmd.Arg(0)
	)
	if len(parsedArgs) > 1 {
		runCmd = strslice.StrSlice(parsedArgs[1:])
	}
	if *flEntrypoint != "" {
		entrypoint = strslice.StrSlice{*flEntrypoint}
	}

	if err := cli.pullImage(context.Background(), image); err != nil {
		return err
	}

	// collect all the environment variables for the container
	envVariables, err := opts.ReadKVStrings(flEnvFile.GetAll(), flEnv.GetAll())
	if err != nil {
		return err
	}

	// collect all the labels for the container
	labels, err := opts.ReadKVStrings(flLabelsFile.GetAll(), flLabels.GetAll())
	if err != nil {
		return err
	}
	labels = append(labels, fmt.Sprintf("sh_hyper_instancetype=%s", *flContainerSize))
	for _, sg := range flSecurityGroups.GetAll() {
		if sg == "" {
			continue
		}
		labels = append(labels, fmt.Sprintf("sh_hyper_sg_%s=yes", sg))
	}
	if *flNoAutoVolume {
		labels = append(labels, "sh_hyper_noauto_volume=true")
	}

	var (
		domainname string
		hostname   = *flHostname
		parts      = strings.SplitN(hostname, ".", 2)
	)
	if len(parts) > 1 {
		hostname = parts[0]
		domainname = parts[1]
	}
	ports, portBindings, err := nat.ParsePortSpecs(flPublish.GetAll())
	if err != nil {
		return err
	}

	// Merge in exposed ports to the map of published ports
	for _, e := range flExpose.GetAll() {
		if strings.Contains(e, ":") {
			return fmt.Errorf("Invalid port format for --expose: %s", e)
		}
		//support two formats for expose, original format <portnum>/[<proto>] or <startport-endport>/[<proto>]
		proto, port := nat.SplitProtoPort(e)
		//parse the start and end port and create a sequence of ports to expose
		//if expose a port, the start and end port are the same
		start, end, err := nat.ParsePortRange(port)
		if err != nil {
			return fmt.Errorf("Invalid range format for --expose: %s, error: %s", e, err)
		}
		for i := start; i <= end; i++ {
			p, err := nat.NewPort(proto, strconv.FormatUint(i, 10))
			if err != nil {
				return err
			}
			if _, exists := ports[p]; !exists {
				ports[p] = struct{}{}
			}
		}
	}
	var binds []string
	// add any bind targets to the list of container volumes
	for bind := range flVolumes.GetMap() {
		if arr := opts.VolumeSplitN(bind, 2); len(arr) > 1 {
			// after creating the bind mount we want to delete it from the flVolumes values because
			// we do not want bind mounts being committed to image configs
			binds = append(binds, bind)
			flVolumes.Delete(bind)
		}
	}
	restartPolicy, err := opts.ParseRestartPolicy(*flRestartPolicy)
	if err != nil {
		return err
	}

	config := &container.Config{
		Hostname:     hostname,
		Domainname:   domainname,
		ExposedPorts: ports,
		Env:          envVariables,
		Cmd:          runCmd,
		Image:        image,
		Volumes:      flVolumes.GetMap(),
		Entrypoint:   entrypoint,
		WorkingDir:   *flWorkingDir,
		Labels:       opts.ConvertKVStringsToMap(labels),
		StopSignal:   *flStopSignal,
	}

	hostConfig := &container.HostConfig{
		Binds:           binds,
		PortBindings:    portBindings,
		Links:           flLinks.GetAll(),
		PublishAllPorts: *flPublishAll,
		NetworkMode:     container.NetworkMode(*flNetMode),
		RestartPolicy:   restartPolicy,
	}
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: make(map[string]*network.EndpointSettings),
	}

	if hostConfig.NetworkMode.IsUserDefined() && len(hostConfig.Links) > 0 {
		epConfig := networkingConfig.EndpointsConfig[string(hostConfig.NetworkMode)]
		if epConfig == nil {
			epConfig = &network.EndpointSettings{}
		}
		epConfig.Links = make([]string, len(hostConfig.Links))
		copy(epConfig.Links, hostConfig.Links)
		networkingConfig.EndpointsConfig[string(hostConfig.NetworkMode)] = epConfig
	}

	if *flMinute == "0" && *flHour == "0" && *flDom == "*" && *flDow == "*" && *flMonth == "*" {
		return fmt.Errorf("must specify at least one schedule")
	}

	sv := types.Cron{
		ContainerName: *flContainerName,
		Schedule:      *flMinute + " " + *flHour + " " + *flDom + " " + *flMonth + " " + *flDow,
		OwnerEmail:    *flMailTo,
		Config:        config,
		HostConfig:    hostConfig,
		NetConfig:     networkingConfig,
	}

	_, err = cli.client.CronCreate(context.Background(), *flName, sv)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "Cron %s is created.\n", *flName)
	return nil
}

// CmdCronDelete deletes one or more crons
//
// Usage: hyper cron rm cron [cron...]
func (cli *DockerCli) CmdCronRm(args ...string) error {
	cmd := Cli.Subcmd("cron rm", []string{"cron [cron...]"}, "Remove one or more crons", false)
	cmd.Require(flag.Min, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	status := 0
	for _, sn := range cmd.Args() {
		if err := cli.client.CronDelete(context.Background(), sn); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
		fmt.Fprintf(cli.out, "%s\n", sn)
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

// CmdCronLs lists all the crons
//
// Usage: hyper cron ls [OPTIONS]
func (cli *DockerCli) CmdCronLs(args ...string) error {
	cmd := Cli.Subcmd("cron ls", nil, "Lists crons", true)

	flFilter := ropts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process after get response from server.
	cronFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		if cronFilterArgs, err = filters.ParseFlag(f, cronFilterArgs); err != nil {
			return err
		}
	}

	options := types.CronListOptions{
		Filters: cronFilterArgs,
	}

	crons, err := cli.client.CronList(context.Background(), options)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprintf(w, "Name\tSchedule\tContainers\tStatus\n")
	for _, cron := range crons {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d/%d\n", cron.Name, cron.Schedule, cron.ContainerName, cron.SuccessCount, cron.ErrorCount)
	}

	w.Flush()
	return nil
}

// CmdCronInspect
//
// Usage: docker cron inspect [OPTIONS] CRON [CRON...]
func (cli *DockerCli) CmdCronInspect(args ...string) error {
	cmd := Cli.Subcmd("cron inspect", []string{"cron [cron...]"}, "Display detailed information on the given cron", true)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	ctx := context.Background()

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		i, err := cli.client.CronInspect(ctx, name)
		return i, nil, err
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

func cronUsage() string {
	cronCommands := [][]string{
		{"create", "Create a cron"},
		{"inspect", "Display detailed information on the given cron"},
		{"ls", "List all crons"},
		{"rm", "Remove one or more crons"},
	}

	help := "Commands:\n"

	for _, cmd := range cronCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	help += fmt.Sprintf("\nRun 'hyper cron COMMAND --help' for more information on a command.")
	return help
}
