package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"

	Cli "github.com/hyperhq/hypercli/cli"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	"github.com/hyperhq/hypercli/pkg/stringid"

	"github.com/cheggaaa/pb"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/hyperhq/hypercli/opts"
	"golang.org/x/net/context"
)

// CmdVolume is the parent subcommand for all volume commands
//
// Usage: docker volume <COMMAND> <OPTS>
func (cli *DockerCli) CmdVolume(args ...string) error {
	description := Cli.DockerCommands["volume"].Description + "\n\nCommands:\n"
	commands := [][]string{
		{"create", "Create a volume"},
		{"inspect", "Return low-level information on a volume"},
		{"ls", "List volumes"},
		{"init", "Initialize volumes"},
		{"rm", "Remove a volume"},
	}

	for _, cmd := range commands {
		description += fmt.Sprintf("  %-25.25s%s\n", cmd[0], cmd[1])
	}

	description += "\nRun 'hyper volume COMMAND --help' for more information on a command"
	cmd := Cli.Subcmd("volume", []string{"[COMMAND]"}, description, false)

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdVolumeLs outputs a list of Docker volumes.
//
// Usage: docker volume ls [OPTIONS]
func (cli *DockerCli) CmdVolumeLs(args ...string) error {
	cmd := Cli.Subcmd("volume ls", nil, "List volumes", true)

	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display volume names")
	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Provide filter values (i.e. 'dangling=true')")

	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	volFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		var err error
		volFilterArgs, err = filters.ParseFlag(f, volFilterArgs)
		if err != nil {
			return err
		}
	}

	volumes, err := cli.client.VolumeList(context.Background(), volFilterArgs)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	if !*quiet {
		for _, warn := range volumes.Warnings {
			fmt.Fprintln(cli.err, warn)
		}
		fmt.Fprintf(w, "DRIVER \tVOLUME NAME\tSIZE\tCONTAINER")
		fmt.Fprintf(w, "\n")
	}

	for _, vol := range volumes.Volumes {
		if *quiet {
			fmt.Fprintln(w, vol.Name)
			continue
		}
		var size, container string
		if vol.Labels != nil {
			size = vol.Labels["size"]
			container = vol.Labels["container"]
			if container != "" {
				container = stringid.TruncateID(container)
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s GB\t%s\n", vol.Driver, vol.Name, size, container)
	}
	w.Flush()
	return nil
}

// CmdVolumeInspect displays low-level information on one or more volumes.
//
// Usage: docker volume inspect [OPTIONS] VOLUME [VOLUME...]
func (cli *DockerCli) CmdVolumeInspect(args ...string) error {
	cmd := Cli.Subcmd("volume inspect", []string{"VOLUME [VOLUME...]"}, "Return low-level information on a volume", true)
	tmplStr := cmd.String([]string{"f", "-format"}, "", "Format the output using the given go template")

	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	ctx := context.Background()

	inspectSearcher := func(name string) (interface{}, []byte, error) {
		i, err := cli.client.VolumeInspect(ctx, name)
		return i, nil, err
	}

	return cli.inspectElements(*tmplStr, cmd.Args(), inspectSearcher)
}

// CmdVolumeCreate creates a new volume.
//
// Usage: docker volume create [OPTIONS]
func (cli *DockerCli) CmdVolumeCreate(args ...string) error {
	cmd := Cli.Subcmd("volume create", nil, "Create a volume", true)
	flDriver := cmd.String([]string{}, "hyper", "Specify volume driver name")
	flName := cmd.String([]string{"-name"}, "", "Specify volume name")
	flSnapshot := cmd.String([]string{"-snapshot"}, "", "Specify snapshot to create volume")
	flSize := cmd.Int([]string{"-size"}, 10, "Specify volume size")

	flDriverOpts := opts.NewMapOpts(nil, nil)
	cmd.Var(flDriverOpts, []string{"o", "-opt"}, "Set driver specific options")

	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	volReq := types.VolumeCreateRequest{
		Driver:     *flDriver,
		DriverOpts: flDriverOpts.GetAll(),
		Name:       *flName,
	}

	volReq.DriverOpts["size"] = fmt.Sprintf("%d", *flSize)
	if *flSnapshot != "" {
		volReq.DriverOpts["snapshot"] = *flSnapshot
		if *flSize == 10 {
			volReq.DriverOpts["size"] = ""
		}
	}

	vol, err := cli.client.VolumeCreate(context.Background(), volReq)
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.out, "%s\n", vol.Name)
	return nil
}

// CmdVolumeRm removes one or more volumes.
//
// Usage: docker volume rm VOLUME [VOLUME...]
func (cli *DockerCli) CmdVolumeRm(args ...string) error {
	cmd := Cli.Subcmd("volume rm", []string{"VOLUME [VOLUME...]"}, "Remove a volume", true)
	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	var status = 0
	ctx := context.Background()
	for _, name := range cmd.Args() {
		if err := cli.client.VolumeRemove(ctx, name); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
		fmt.Fprintf(cli.out, "%s\n", name)
	}

	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

func validateVolumeSource(source string) error {
	switch {
	case strings.HasPrefix(source, "git://"):
		fallthrough
	case strings.HasPrefix(source, "http://"):
		fallthrough
	case strings.HasPrefix(source, "https://"):
		break
	case strings.HasPrefix(source, "/"):
		info, err := os.Stat(source)
		if err != nil {
			return err
		}
		if !info.Mode().IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("Unsupported local volume source(%s): %s", source, info.Mode().String())
		}
		break
	default:
		return fmt.Errorf("%s is not supported volume source", source)
	}

	return nil
}

func validateVolumeInitArgs(args []string, req *types.VolumesInitializeRequest) error {

	for _, desc := range args {
		idx := strings.LastIndexByte(desc, ':')
		if idx == -1 || idx >= len(desc)-1 {
			return fmt.Errorf("%s does not match format SOURCE:VOLUME", desc)
		}
		source := desc[:idx]
		name := desc[idx+1:]
		if err := validateVolumeSource(source); err != nil {
			return err
		}
		req.Volume = append(req.Volume, types.VolumeInitDesc{
			Name:   name,
			Source: source,
		})
	}
	return nil
}

// CmdVolumeInit Initializes one or more volumes.
//
// Usage: docker volume init SOURCE:VOLUME [SOURCE:VOLUME...]
func (cli *DockerCli) CmdVolumeInit(args ...string) error {
	cmd := Cli.Subcmd("volume init", []string{"SOURCE:VOLUME [SOURCE:VOLUME...]"}, "Initialize a volume", true)
	cmd.Require(flag.Min, 1)
	cmd.ParseFlags(args, true)

	return cli.initVolumes(cmd.Args(), false)
}

func (cli *DockerCli) initVolumes(vols []string, reload bool) error {
	var req types.VolumesInitializeRequest
	err := validateVolumeInitArgs(vols, &req)
	if err != nil {
		return err
	}

	ctx := context.Background()
	req.Reload = reload
	resp, err := cli.client.VolumeInitialize(ctx, req)
	if err != nil {
		return err
	}

	if len(resp.Session) == 0 {
		return nil
	}

	// Upload local volumes
	var wg sync.WaitGroup
	var results []error
	pool, err := pb.StartPool()
	if err != nil {
		// Ignore progress bar failures
		fmt.Fprintf(cli.err, "Warning: do not show upload progress: %s\n", err.Error())
		pool = nil
		err = nil
	}
	for _, desc := range req.Volume {
		if url, ok := resp.Uploaders[desc.Name]; ok {
			wg.Add(1)
			go uploadLocalVolume(desc.Source, url, resp.Cookie, &results, &wg, pool)
		}
	}

	wg.Wait()
	if pool != nil {
		pool.Stop()
	}
	for _, err = range results {
		fmt.Fprintf(cli.err, "Upload local volume failed: %s\n", err.Error())
	}

	finishErr := cli.client.VolumeUploadFinish(ctx, resp.Session)
	if err == nil {
		err = finishErr
	}
	return err
}

func uploadLocalVolume(source, url, cookie string, results *[]error, wg *sync.WaitGroup, pool *pb.Pool) {
	var (
		resp     io.ReadCloser
		tar      *TarFile
		fullPath string
		err      error
	)

	defer func() {
		if err != nil {
			*results = append(*results, err)
		}
		wg.Done()
	}()

	fullPath, err = filepath.Abs(source)
	if err != nil {
		return
	}

	tar = NewTarFile(source, 512)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		var relPath, linkName string

		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkName, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		if path == fullPath {
			if info.IsDir() {
				// "." as indicator that it is a dir volume
				relPath = "."
			} else {
				relPath = filepath.Base(path)
			}
		} else {
			relPath, err = filepath.Rel(fullPath, path)
			if err != nil {
				return err
			}
		}
		tar.AddFile(info, relPath, linkName, path)
		return nil
	}

	err = filepath.Walk(fullPath, walkFunc)
	if err != nil {
		return
	}
	if pool != nil {
		tar.AllocBar(pool)
	}

	resp, err = sendTarball(url, cookie, tar)
	if err != nil {
		return
	}
	defer resp.Close()
}

func sendTarball(uri, cookie string, input io.ReadCloser) (io.ReadCloser, error) {
	req, err := http.NewRequest("POST", uri+"?cookie="+cookie, input)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-tar")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		if buf.Len() > 0 {
			err = fmt.Errorf("%s: %s", http.StatusText(resp.StatusCode), buf.String())
		} else {
			err = fmt.Errorf("%s", http.StatusText(resp.StatusCode))
		}
		return nil, err
	}
	return resp.Body, nil
}
