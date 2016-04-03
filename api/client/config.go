package client

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/cliconfig"
	flag "github.com/docker/docker/pkg/mflag"
)

// CmdConfig
//
// Usage: hyper config
func (cli *DockerCli) CmdConfig(args ...string) error {
	cmd := Cli.Subcmd("config", []string{"[SERVER]"}, Cli.DockerCommands["config"].Description+".\nIf no server is specified, the default is defined as "+cliconfig.DefaultHyperServer, true)
	cmd.Require(flag.Max, 0)

	flAccesskey := cmd.String([]string{"-accesskey"}, "", "Access Key")
	flSecretkey := cmd.String([]string{"-secretkey"}, "", "Secret Key")

	cmd.ParseFlags(args, true)

	// On Windows, force the use of the regular OS stdin stream. Fixes #14336/#14210
	if runtime.GOOS == "windows" {
		cli.in = os.Stdin
	}

	var serverAddress string
	if len(cmd.Args()) > 0 {
		serverAddress = cmd.Arg(0)
	} else {
		serverAddress = cliconfig.DefaultHyperServer
	}

	_, err := cli.configureCloud(serverAddress, *flAccesskey, *flSecretkey)
	if err != nil {
		return err
	}

	if err := cli.configFile.Save(); err != nil {
		return fmt.Errorf("Error saving config file: %v", err)
	}
	fmt.Fprintf(cli.out, "WARNING: Your login credentials has been saved in %s\n", cli.configFile.Filename())

	return nil
}

func (cli *DockerCli) configureCloud(serverAddress, flAccesskey, flSecretkey string) (cliconfig.CloudConfig, error) {
	cloudConfig, ok := cli.configFile.CloudConfig[serverAddress]
	if !ok {
		cloudConfig = cliconfig.CloudConfig{}
	}

	if flAccesskey = strings.TrimSpace(flAccesskey); flAccesskey == "" {
		cli.promptWithDefault("Enter Access Key", cloudConfig.AccessKey)
		flAccesskey = readInput(cli.in, cli.out)
		flAccesskey = strings.TrimSpace(flAccesskey)
	}
	if flSecretkey = strings.TrimSpace(flSecretkey); flSecretkey == "" {
		cli.promptWithDefault("Enter Secret Key", cloudConfig.SecretKey)
		flSecretkey = readInput(cli.in, cli.out)
		flSecretkey = strings.TrimSpace(flSecretkey)
	}

	cloudConfig.AccessKey = flAccesskey
	cloudConfig.SecretKey = flSecretkey
	cli.configFile.CloudConfig[serverAddress] = cloudConfig
	return cloudConfig, nil
}
