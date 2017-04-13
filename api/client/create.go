package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	networktypes "github.com/docker/engine-api/types/network"
	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/pkg/jsonmessage"
	"github.com/hyperhq/hypercli/reference"
	"github.com/hyperhq/hypercli/registry"
	runconfigopts "github.com/hyperhq/hypercli/runconfig/opts"
)

func (cli *DockerCli) pullImage(ctx context.Context, image string) error {
	return cli.pullImageCustomOut(ctx, image, cli.out)
}

func (cli *DockerCli) pullImageCustomOut(ctx context.Context, image string, out io.Writer) error {
	ref, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	authConfig := cli.resolveAuthConfig(ctx, cli.configFile.AuthConfigs, repoInfo.Index)
	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	options := types.ImageCreateOptions{
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.client.ImageCreate(ctx, image, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, out, cli.outFd, cli.isTerminalOut, nil)
}

type cidFile struct {
	path    string
	file    *os.File
	written bool
}

func newCIDFile(path string) (*cidFile, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("Container ID file found, make sure the other container isn't running or delete %s", path)
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create the container ID file: %s", err)
	}

	return &cidFile{path: path, file: f}, nil
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
	case filepath.VolumeName(bind) != "":
		// Windows local path
	default:
		return "", "", false
	}

	pos := strings.LastIndex(bind, ":")
	if pos < 0 || pos >= len(bind)-1 {
		return "", "", false
	}

	return bind[:pos], bind[pos+1:], true
}

func (cli *DockerCli) createContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, cidfile, name string) (*types.ContainerCreateResponse, error) {
	var containerIDFile *cidFile
	var initvols, volumeList []string

	if cidfile != "" {
		var err error
		if containerIDFile, err = newCIDFile(cidfile); err != nil {
			return nil, err
		}
		defer containerIDFile.Close()
	}

	// Check/create protocol and local volume
	defer func() {
		for _, vol := range volumeList {
			cli.client.VolumeRemove(ctx, vol)
		}
	}()
	for idx, bind := range hostConfig.Binds {
		if source, dest, ok := parseProtoAndLocalBind(bind); ok {
			volReq := types.VolumeCreateRequest{
				Driver: "hyper",
				Labels: map[string]string{
					"autoremove": "true",
				}}
			if vol, err := cli.client.VolumeCreate(ctx, volReq); err != nil {
				return nil, err
			} else {
				initvols = append(initvols, source+":"+vol.Name)
				volumeList = append(volumeList, vol.Name)
				hostConfig.Binds[idx] = vol.Name + ":" + dest
			}
		}
	}

	// initialize special volumes
	if len(initvols) > 0 {
		err := cli.initVolumes(initvols, false)
		if err != nil {
			return nil, err
		}
	}

	ref, err := reference.ParseNamed(config.Image)
	if err != nil {
		return nil, err
	}
	ref = reference.WithDefaultTag(ref)

	var trustedRef reference.Canonical

	if ref, ok := ref.(reference.NamedTagged); ok && isTrusted() {
		var err error
		trustedRef, err = cli.trustedReference(ctx, ref)
		if err != nil {
			return nil, err
		}
		config.Image = trustedRef.String()
	}

	//create the container
	response, err := cli.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, name)

	//if image not found try to pull it
	if err != nil {
		if client.IsErrImageNotFound(err) {
			fmt.Fprintf(cli.err, "Unable to find image '%s' in the current region\n", ref.String())

			// we don't want to write to stdout anything apart from container.ID
			if err = cli.pullImageCustomOut(ctx, config.Image, cli.err); err != nil {
				return nil, err
			}
			if ref, ok := ref.(reference.NamedTagged); ok && trustedRef != nil {
				if err := cli.tagTrusted(ctx, trustedRef, ref); err != nil {
					return nil, err
				}
			}
			// Retry
			var retryErr error
			response, retryErr = cli.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, name)
			if retryErr != nil {
				return nil, retryErr
			}
		} else {
			return nil, err
		}
	}
	volumeList = nil

	for _, warning := range response.Warnings {
		fmt.Fprintf(cli.err, "WARNING: %s\n", warning)
	}
	if containerIDFile != nil {
		if err = containerIDFile.Write(response.ID); err != nil {
			return nil, err
		}
	}
	return &response, nil
}

// CmdCreate creates a new container from a given image.
//
// Usage: docker create [OPTIONS] IMAGE [COMMAND] [ARG...]
func (cli *DockerCli) CmdCreate(args ...string) error {
	cmd := Cli.Subcmd("create", []string{"IMAGE [COMMAND] [ARG...]"}, Cli.DockerCommands["create"].Description, true)
	addTrustedFlags(cmd, true)

	// These are flags not stored in Config/HostConfig
	var (
		flName = cmd.String([]string{"-name"}, "", "Assign a name to the container")
	)

	config, hostConfig, networkingConfig, cmd, err := runconfigopts.Parse(cmd, args)

	if err != nil {
		cmd.ReportError(err.Error(), true)
		os.Exit(1)
	}
	if config.Image == "" {
		cmd.Usage()
		return nil
	}
	response, err := cli.createContainer(context.Background(), config, hostConfig, networkingConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s\n", response.ID)
	return nil
}
