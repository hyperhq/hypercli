package client

import (
	"io"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/pkg/jsonmessage"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	"github.com/hyperhq/hypercli/reference"
	"github.com/hyperhq/hypercli/registry"
)

// CmdPush pushes an image or repository to the registry.
//
// Usage: docker push NAME[:TAG]
func (cli *DockerCli) Push(args ...string) error {
	cmd := Cli.Subcmd("push", []string{"NAME[:TAG]"}, Cli.DockerCommands["push"].Description, true)
	addTrustedFlags(cmd, false)
	cmd.Require(flag.Exact, 1)

	cmd.ParseFlags(args, true)

	ref, err := reference.ParseNamed(cmd.Arg(0))
	if err != nil {
		return err
	}
	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Resolve the Auth config relevant for this server
	authConfig := cli.resolveAuthConfig(ctx, cli.configFile.AuthConfigs, repoInfo.Index)
	requestPrivilege := cli.registryAuthenticationPrivilegedFunc(repoInfo.Index, "push")

	if isTrusted() {
		return cli.trustedPush(ctx, repoInfo, ref, authConfig, requestPrivilege)
	}

	responseBody, err := cli.imagePushPrivileged(ctx, authConfig, ref.String(), requestPrivilege)
	if err != nil {
		return err
	}

	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}

func (cli *DockerCli) imagePushPrivileged(ctx context.Context, authConfig types.AuthConfig, ref string, requestPrivilege types.RequestPrivilegeFunc) (io.ReadCloser, error) {
	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return nil, err
	}
	options := types.ImagePushOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
	}

	return cli.client.ImagePush(ctx, ref, options)
}
