package client

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	networktypes "github.com/docker/engine-api/types/network"
	"github.com/docker/libnetwork/resolvconf/dns"
	Cli "github.com/hyperhq/hypercli/cli"
	derr "github.com/hyperhq/hypercli/errors"
	"github.com/hyperhq/hypercli/opts"
	"github.com/hyperhq/hypercli/pkg/promise"
	"github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/pkg/stringid"
	runconfigopts "github.com/hyperhq/hypercli/runconfig/opts"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

type InitVolume struct {
	Source      string
	Destination string
	Name        string
}

func (cid *cidFile) Close() error {
	cid.file.Close()

	if !cid.written {
		if err := os.Remove(cid.path); err != nil {
			return fmt.Errorf("failed to remove the CID file '%s': %s \n", cid.path, err)
		}
	}

	return nil
}

func (cid *cidFile) Write(id string) error {
	if _, err := cid.file.Write([]byte(id)); err != nil {
		return fmt.Errorf("Failed to write the container ID to the file: %s", err)
	}
	cid.written = true
	return nil
}

// if container start fails with 'command not found' error, return 127
// if container start fails with 'command cannot be invoked' error, return 126
// return 125 for generic docker daemon failures
func runStartContainerErr(err error) error {
	trimmedErr := strings.Trim(err.Error(), "Error response from daemon: ")
	statusError := Cli.StatusError{}
	derrCmdNotFound := derr.ErrorCodeCmdNotFound.Message()
	derrCouldNotInvoke := derr.ErrorCodeCmdCouldNotBeInvoked.Message()
	derrNoSuchImage := derr.ErrorCodeNoSuchImageHash.Message()
	derrNoSuchImageTag := derr.ErrorCodeNoSuchImageTag.Message()
	switch trimmedErr {
	case derrCmdNotFound:
		statusError = Cli.StatusError{StatusCode: 127}
	case derrCouldNotInvoke:
		statusError = Cli.StatusError{StatusCode: 126}
	case derrNoSuchImage, derrNoSuchImageTag:
		statusError = Cli.StatusError{StatusCode: 125}
	default:
		statusError = Cli.StatusError{StatusCode: 125}
	}
	return statusError
}

func parseProtoAndLocalBind(bind string) (string, string, bool) {
	switch {
	case strings.HasPrefix(bind, "git://"):
		fallthrough
	case strings.HasPrefix(bind, "http://"):
		fallthrough
	case strings.HasPrefix(bind, "https://"):
		if strings.Count(bind, ":") < 2 {
			return "", "", false
		}
	case strings.HasPrefix(bind, "/"):
		if strings.Count(bind, ":") < 1 {
			return "", "", false
		}
	default:
		return "", "", false
	}

	pos := strings.LastIndex(bind, ":")
	if pos < 0 || pos >= len(bind)-1 {
		return "", "", false
	}

	return bind[:pos], bind[pos+1:], true
}

func checkSourceType(source string) string {
	part := strings.Split(source, ":")
	count := len(part)
	switch {
	case strings.HasPrefix(source, "git://") || strings.HasSuffix(source, ".git") ||
		(count >= 2 && strings.HasSuffix(part[count-2], ".git")):
		return "git"
	case strings.HasPrefix(source, "http://"):
		fallthrough
	case strings.HasPrefix(source, "https://"):
		return "http"
	case strings.HasPrefix(source, "/"):
		return "local"
	default:
		return "unknown"
	}
}

func (cli *DockerCli) initSpecialVolumes(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, initvols []*InitVolume) error {
	const INIT_VOLUME_PATH = "/vol/"
	const INIT_VOLUME_IMAGE = "hyperhq/volume_uploader:v1"
	const INIT_VOLUME_FILENAME = ".hyper_file_volume_data_do_not_create_on_your_own"
	var (
		initConfig     *container.Config
		initHostConfig *container.HostConfig
		errCh          chan error
		execCount      uint32
		fip            string
	)

	initConfig = &container.Config{
		User:       config.User,
		Image:      INIT_VOLUME_IMAGE,
		StopSignal: config.StopSignal,
	}

	initHostConfig = &container.HostConfig{
		Binds:      make([]string, 0),
		DNS:        hostConfig.DNS,
		DNSOptions: hostConfig.DNSOptions,
		DNSSearch:  hostConfig.DNSSearch,
		ExtraHosts: hostConfig.ExtraHosts,
	}

	for _, vol := range initvols {
		initHostConfig.Binds = append(initHostConfig.Binds, vol.Name+":"+INIT_VOLUME_PATH+vol.Destination)
	}
	passwd := uuid.NewV1()
	initConfig.Env = append(config.Env, "ROOTPASSWORD="+passwd.String(), "LOCALROOT="+INIT_VOLUME_PATH)

	createResponse, err := cli.createContainer(ctx, initConfig, initHostConfig, networkingConfig, hostConfig.ContainerIDFile, "")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if _, rmErr := cli.removeContainer(ctx, createResponse.ID, true, false, true); rmErr != nil {
				fmt.Fprintf(cli.err, "clean up init container failed: %s\n", rmErr.Error())
			}
		}
		if fip != "" {
			if rmErr := cli.releaseFip(ctx, fip); rmErr != nil {
				fmt.Fprintf(cli.err, "failed to clean up container fip %s: %s\n", fip, rmErr.Error())
			}
		}
	}()

	if err = cli.client.ContainerStart(ctx, createResponse.ID, ""); err != nil {
		return err
	}

	errCh = make(chan error, len(initvols))
	for _, vol := range initvols {
		var cmd []string
		volType := checkSourceType(vol.Source)
		switch volType {
		case "git":
			cmd = append(cmd, "git", "clone", vol.Source, INIT_VOLUME_PATH+vol.Destination)
		case "http":
			isFile, err := fileSourceVolume(vol.Source)
			if err != nil {
				return err
			}
			parts := strings.Split(vol.Source, "/")
			if isFile {
				cmd = append(cmd, "wget", "--no-check-certificate", "--tries=5", vol.Source, "--output-document="+INIT_VOLUME_PATH+vol.Destination+"/"+INIT_VOLUME_FILENAME)
			} else {
				cmd = append(cmd, "wget", "--no-check-certificate", "--tries=5", "--mirror", "--no-host-directories", "--cut-dirs="+strconv.Itoa(len(parts)), vol.Source, "--directory-prefix="+INIT_VOLUME_PATH+vol.Destination)
			}
		case "local":
			isFile, err := fileSourceVolume(vol.Source)
			if err != nil {
				return err
			}
			execCount++
			if fip == "" {
				fip, err = cli.associateNewFip(ctx, createResponse.ID)
				if err != nil {
					return err
				}
			}
			go func(vol *InitVolume) {
				dest := INIT_VOLUME_PATH + vol.Destination
				if isFile {
					dest = dest + "/" + INIT_VOLUME_FILENAME
				}
				err := cli.uploadLocalResource(vol.Source, dest, fip, "root", passwd.String(), isFile)
				if err != nil {
					err = fmt.Errorf("Failed to upload %s: %s", vol.Source, err.Error())
				}
				errCh <- err
			}(vol)
		default:
			continue
		}
		if len(cmd) == 0 {
			continue
		}

		execCount++
		go func() {
			execID, err := cli.ExecCmd(ctx, initConfig.User, createResponse.ID, cmd)
			if err != nil {
				errCh <- err
			} else {
				errCh <- cli.WaitExec(ctx, execID)
			}
		}()
	}

	// wait results
	for ; execCount > 0; execCount-- {
		err = <-errCh
		if err != nil {
			return err
		}
	}

	// release fip
	if fip != "" {
		if err = cli.releaseContainerFip(ctx, createResponse.ID); err != nil {
			return err
		}
		fip = ""
	}

	// Need to sync before tearing down container because data might be still cached
	syncCmd := []string{"sync"}
	execID, err := cli.ExecCmd(ctx, initConfig.User, createResponse.ID, syncCmd)
	if err != nil {
		return err
	}
	if err = cli.WaitExec(ctx, execID); err != nil {
		return err
	}

	_, err = cli.removeContainer(ctx, createResponse.ID, false, false, true)
	if err != nil {
		return err
	}

	return nil
}

func fileSourceVolume(source string) (bool, error) {
	switch {
	case strings.HasPrefix(source, "git://"):
		return false, nil
	case strings.HasPrefix(source, "http://"):
		fallthrough
	case strings.HasPrefix(source, "https://"):
		part := strings.Split(source, ":")
		count := len(part)
		if strings.HasSuffix(source, "/") || strings.HasSuffix(source, ".git") ||
			(count >= 3 && strings.HasSuffix(part[count-2], ".git")) {
			return false, nil
		} else {
			return true, nil
		}
	case strings.HasPrefix(source, "/"):
		info, err := os.Stat(source)
		if err != nil {
			return false, err
		}
		if info.IsDir() {
			return false, nil
		} else if info.Mode()&os.ModeType != 0 {
			return false, fmt.Errorf("cannot init volume from special file: %s", source)
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported volume source type: %s", source)
	}

}

// CmdRun runs a command in a new container.
//
// Usage: docker run [OPTIONS] IMAGE [COMMAND] [ARG...]
func (cli *DockerCli) CmdRun(args ...string) error {
	cmd := Cli.Subcmd("run", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["run"].Description, true)
	addTrustedFlags(cmd, true)

	// These are flags not stored in Config/HostConfig
	var (
		flAutoRemove = cmd.Bool([]string{"-rm"}, false, "Automatically remove the container when it exits")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Run container in background and print container ID")
		flSigProxy   = cmd.Bool([]string{}, true, "Proxy received signals to the process")
		flName       = cmd.String([]string{"-name"}, "", "Assign a name to the container")
		flDetachKeys = cmd.String([]string{}, "", "Override the key sequence for detaching a container")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)

	// just in case the Parse does not exit
	if err != nil {
		cmd.ReportError(err.Error(), true)
		os.Exit(125)
	}

	if hostConfig.OomKillDisable != nil && *hostConfig.OomKillDisable && hostConfig.Memory == 0 {
		fmt.Fprintf(cli.err, "WARNING: Disabling the OOM killer on containers without setting a '-m/--memory' limit may be dangerous.\n")
	}

	if len(hostConfig.DNS) > 0 {
		// check the DNS settings passed via --dns against
		// localhost regexp to warn if they are trying to
		// set a DNS to a localhost address
		for _, dnsIP := range hostConfig.DNS {
			if dns.IsLocalhost(dnsIP) {
				fmt.Fprintf(cli.err, "WARNING: Localhost DNS setting (--dns=%s) may fail in containers.\n", dnsIP)
				break
			}
		}
	}
	if config.Image == "" {
		cmd.Usage()
		return nil
	}

	config.ArgsEscaped = false

	if !*flDetach {
		if err := cli.CheckTtyInput(config.AttachStdin, config.Tty); err != nil {
			return err
		}
	} else {
		if fl := cmd.Lookup("-attach"); fl != nil {
			flAttach = fl.Value.(*opts.ListOpts)
			if flAttach.Len() != 0 {
				return ErrConflictAttachDetach
			}
		}
		if *flAutoRemove {
			return ErrConflictDetachAutoRemove
		}

		config.AttachStdin = false
		config.AttachStdout = false
		config.AttachStderr = false
		config.StdinOnce = false
	}

	// Disable flSigProxy when in TTY mode
	sigProxy := *flSigProxy
	if config.Tty {
		sigProxy = false
	}

	// Telling the Windows daemon the initial size of the tty during start makes
	// a far better user experience rather than relying on subsequent resizes
	// to cause things to catch up.
	if runtime.GOOS == "windows" {
		hostConfig.ConsoleSize[0], hostConfig.ConsoleSize[1] = cli.getTtySize()
	}

	ctx := context.Background()

	// Check/create protocol and local volume
	var initvols []*InitVolume
	defer func() {
		for _, vol := range initvols {
			cli.client.VolumeRemove(ctx, vol.Name)
		}
	}()
	for idx, bind := range hostConfig.Binds {
		if source, dest, ok := parseProtoAndLocalBind(bind); ok {
			volReq := types.VolumeCreateRequest{
				Driver: "hyper",
				Labels: map[string]string{
					"source": source,
				}}
			if vol, err := cli.client.VolumeCreate(ctx, volReq); err != nil {
				cmd.ReportError(err.Error(), true)
				return runStartContainerErr(err)
			} else {
				initvols = append(initvols, &InitVolume{
					Source:      source,
					Destination: dest,
					Name:        vol.Name,
				})
				hostConfig.Binds[idx] = vol.Name + ":" + dest
			}
		}
	}

	// initialize special volumes
	if len(initvols) > 0 {
		err := cli.initSpecialVolumes(ctx, config, hostConfig, networkingConfig, initvols)
		if err != nil {
			cmd.ReportError(err.Error(), true)
			return runStartContainerErr(err)
		}
	}

	createResponse, err := cli.createContainer(ctx, config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		cmd.ReportError(err.Error(), true)
		return runStartContainerErr(err)
	}
	initvols = nil

	if sigProxy {
		sigc := cli.forwardAllSignals(ctx, createResponse.ID)
		defer signal.StopCatch(sigc)
	}
	var (
		waitDisplayID chan struct{}
		errCh         chan error
	)
	if !config.AttachStdout && !config.AttachStderr {
		// Make this asynchronous to allow the client to write to stdin before having to read the ID
		waitDisplayID = make(chan struct{})
		go func() {
			defer close(waitDisplayID)
			fmt.Fprintf(cli.out, "%s\n", createResponse.ID)
		}()
	}
	if *flAutoRemove && (hostConfig.RestartPolicy.IsAlways() || hostConfig.RestartPolicy.IsOnFailure()) {
		return ErrConflictRestartPolicyAndAutoRemove
	}

	if config.AttachStdin || config.AttachStdout || config.AttachStderr {
		var (
			out, stderr io.Writer
			in          io.ReadCloser
		)
		if config.AttachStdin {
			in = cli.in
		}
		if config.AttachStdout {
			out = cli.out
		}
		if config.AttachStderr {
			if config.Tty {
				stderr = cli.out
			} else {
				stderr = cli.err
			}
		}

		if *flDetachKeys != "" {
			cli.configFile.DetachKeys = *flDetachKeys
		}

		options := types.ContainerAttachOptions{
			Stream:     true,
			Stdin:      config.AttachStdin,
			Stdout:     config.AttachStdout,
			Stderr:     config.AttachStderr,
			DetachKeys: cli.configFile.DetachKeys,
		}

		resp, err := cli.client.ContainerAttach(ctx, createResponse.ID, options)
		if err != nil {
			return err
		}
		if in != nil && config.Tty {
			if err := cli.setRawTerminal(); err != nil {
				return err
			}
			defer cli.restoreTerminal(in)
		}
		errCh = promise.Go(func() error {
			return cli.holdHijackedConnection(config.Tty, in, out, stderr, resp)
		})
	}

	if *flAutoRemove {
		defer func() {
			if _, err := cli.removeContainer(ctx, createResponse.ID, true, false, false); err != nil {
				fmt.Fprintf(cli.err, "%v\n", err)
			}
		}()
	}

	//start the container
	if err := cli.client.ContainerStart(ctx, createResponse.ID, ""); err != nil {
		cmd.ReportError(err.Error(), false)
		return runStartContainerErr(err)
	}

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		if err := cli.monitorTtySize(ctx, createResponse.ID, false); err != nil {
			fmt.Fprintf(cli.err, "Error monitoring TTY size: %s\n", err)
		}
	}

	if errCh != nil {
		if err := <-errCh; err != nil {
			logrus.Debugf("Error hijack: %s", err)
			return err
		}
	}

	// Detached mode: wait for the id to be displayed and return.
	if !config.AttachStdout && !config.AttachStderr {
		// Detached mode
		<-waitDisplayID
		return nil
	}

	var status int

	// Attached mode
	if *flAutoRemove {
		// Warn user if they detached us
		js, err := cli.client.ContainerInspect(ctx, createResponse.ID)
		if err != nil {
			return runStartContainerErr(err)
		}
		if js.State.Running == true || js.State.Paused == true {
			fmt.Fprintf(cli.out, "Detached from %s, awaiting its termination in order to uphold \"--rm\".\n",
				stringid.TruncateID(createResponse.ID))
		}

		// Autoremove: wait for the container to finish, retrieve
		// the exit code and remove the container
		if status, err = cli.client.ContainerWait(ctx, createResponse.ID); err != nil {
			return runStartContainerErr(err)
		}
		if _, status, err = getExitCode(ctx, cli, createResponse.ID); err != nil {
			return err
		}
	} else {
		// No Autoremove: Simply retrieve the exit code
		if !config.Tty {
			// In non-TTY mode, we can't detach, so we must wait for container exit
			if status, err = cli.client.ContainerWait(ctx, createResponse.ID); err != nil {
				return err
			}
		} else {
			// In TTY mode, there is a race: if the process dies too slowly, the state could
			// be updated after the getExitCode call and result in the wrong exit code being reported
			if _, status, err = getExitCode(ctx, cli, createResponse.ID); err != nil {
				return err
			}
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}
