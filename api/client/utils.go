package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	gosignal "os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	networktypes "github.com/docker/engine-api/types/network"
	registrytypes "github.com/docker/engine-api/types/registry"
	"github.com/dutchcoders/goftp"
	"github.com/hyperhq/hypercli/pkg/signal"
	"github.com/hyperhq/hypercli/pkg/term"
	"github.com/hyperhq/hypercli/registry"
)

func (cli *DockerCli) electAuthServer() string {
	// The daemon `/info` endpoint informs us of the default registry being
	// used. This is essential in cross-platforms environment, where for
	// example a Linux client might be interacting with a Windows daemon, hence
	// the default registry URL might be Windows specific.
	serverAddress := registry.IndexServer
	if info, err := cli.client.Info(); err != nil {
		fmt.Fprintf(cli.out, "Warning: failed to get default registry endpoint from daemon (%v). Using system default: %s\n", err, serverAddress)
	} else {
		serverAddress = info.IndexServerAddress
	}
	return serverAddress
}

// encodeAuthToBase64 serializes the auth configuration as JSON base64 payload
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func (cli *DockerCli) registryAuthenticationPrivilegedFunc(index *registrytypes.IndexInfo, cmdName string) client.RequestPrivilegeFunc {
	return func() (string, error) {
		fmt.Fprintf(cli.out, "\nPlease login prior to %s:\n", cmdName)
		indexServer := registry.GetAuthConfigKey(index)
		authConfig, err := cli.configureAuth("", "", "", indexServer)
		if err != nil {
			return "", err
		}
		return encodeAuthToBase64(authConfig)
	}
}

func (cli *DockerCli) resizeTty(id string, isExec bool) {
	height, width := cli.getTtySize()
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		ID:     id,
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = cli.client.ContainerExecResize(options)
	} else {
		err = cli.client.ContainerResize(options)
	}

	if err != nil {
		logrus.Debugf("Error resize: %s", err)
	}
}

// getExitCode perform an inspect on the container. It returns
// the running state and the exit code.
func getExitCode(cli *DockerCli, containerID string) (bool, int, error) {
	c, err := cli.client.ContainerInspect(containerID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return c.State.Running, c.State.ExitCode, nil
}

// getExecExitCode perform an inspect on the exec command. It returns
// the running state and the exit code.
func getExecExitCode(cli *DockerCli, execID string) (bool, int, error) {
	resp, err := cli.client.ContainerExecInspect(execID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != client.ErrConnectionFailed {
			return false, -1, err
		}
		return false, -1, nil
	}

	return resp.Running, resp.ExitCode, nil
}

func (cli *DockerCli) monitorTtySize(id string, isExec bool) error {
	cli.resizeTty(id, isExec)

	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := cli.getTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := cli.getTtySize()

				if prevW != w || prevH != h {
					cli.resizeTty(id, isExec)
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				cli.resizeTty(id, isExec)
			}
		}()
	}
	return nil
}

func (cli *DockerCli) getTtySize() (int, int) {
	if !cli.isTerminalOut {
		return 0, 0
	}
	ws, err := term.GetWinsize(cli.outFd)
	if err != nil {
		logrus.Debugf("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return int(ws.Height), int(ws.Width)
}

func copyToFile(outfile string, r io.Reader) error {
	tmpFile, err := ioutil.TempFile(filepath.Dir(outfile), ".docker_temp_")
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, r)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err = os.Rename(tmpPath, outfile); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// resolveAuthConfig is like registry.ResolveAuthConfig, but if using the
// default index, it uses the default index name for the daemon's platform,
// not the client's platform.
func (cli *DockerCli) resolveAuthConfig(authConfigs map[string]types.AuthConfig, index *registrytypes.IndexInfo) types.AuthConfig {
	configKey := index.Name
	if index.Official {
		configKey = cli.electAuthServer()
	}

	// First try the happy case
	if c, found := authConfigs[configKey]; found || index.Official {
		return c
	}

	convertToHostname := func(url string) string {
		stripped := url
		if strings.HasPrefix(url, "http://") {
			stripped = strings.Replace(url, "http://", "", 1)
		} else if strings.HasPrefix(url, "https://") {
			stripped = strings.Replace(url, "https://", "", 1)
		}

		nameParts := strings.SplitN(stripped, "/", 2)

		return nameParts[0]
	}

	// Maybe they have a legacy config file, we will iterate the keys converting
	// them to the new format and testing
	for registry, ac := range authConfigs {
		if configKey == convertToHostname(registry) {
			return ac
		}
	}

	// When all else fails, return an empty auth config
	return types.AuthConfig{}
}

// uploadLocalResource upload a local file/directory to a container
func (cli *DockerCli) uploadLocalResource(source, dest, serverIP, user, passwd string) error {
	if !strings.Contains(serverIP, ":") {
		serverIP += ":21"
	}
	ftp, err := goftp.Connect(serverIP)
	if err != nil {
		return err
	}

	defer func() {
		ftp.Close()
	}()

	if err = ftp.Login(user, passwd); err != nil {
		return err
	}

	if err = ftp.Cwd(dest); err != nil {
		return err
	}

	if err = ftp.Upload(source); err != nil {
		return err
	}

	return nil
}

func (cli *DockerCli) containerFilterInitVolumes(containerID string) ([]*InitVolume, error) {
	_, raw, err := cli.client.ContainerInspectWithRaw(containerID, false)
	if err != nil {
		return nil, fmt.Errorf("Failed to inspect container: %s", err.Error())
	}

	var contMap map[string]*json.RawMessage
	err = json.Unmarshal(raw, &contMap)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal inspect result: %s", err.Error())
	}

	var hostMap map[string]*json.RawMessage
	err = json.Unmarshal(*contMap["HostConfig"], &hostMap)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal inspected result: %s", err.Error())
	}

	binds := make([]string, 0)
	err = json.Unmarshal(*hostMap["Binds"], &binds)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal binding result: %s", err.Error())
	}

	var initvols []*InitVolume
	for _, bind := range binds {
		part := strings.Split(bind, ":")
		count := len(part)
		if count != 2 {
			continue
		}
		vol, err := cli.client.VolumeInspect(part[0])
		if err != nil {
			return nil, err
		}
		source, ok := vol.Labels["source"]
		if ok {
			initvols = append(initvols, &InitVolume{
				Source:      source,
				Destination: part[1],
				Name:        vol.Name,
			})
		}
	}

	return initvols, nil
}

func (cli *DockerCli) containerReloadInitVolumes(initvols []*InitVolume) error {
	var (
		config           container.Config
		hostConfig       container.HostConfig
		networkingConfig networktypes.NetworkingConfig
	)

	config.StopSignal = "SIGTERM"
	err := cli.initSpecialVolumes(&config, &hostConfig, &networkingConfig, initvols)
	if err != nil {
		return err
	}

	return nil
}
