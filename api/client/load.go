package client

import (
	"io"
	"os"

	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/pkg/jsonmessage"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

// CmdLoad loads an image from a tar archive.
//
// The tar archive is read from STDIN by default, or from a tar archive file.
//
// Usage: docker load [OPTIONS]
func (cli *DockerCli) Load(args ...string) error {
	cmd := Cli.Subcmd("load", nil, Cli.DockerCommands["load"].Description, true)
	infile := cmd.String([]string{"i", "-input"}, "", "Read from a tar archive file, instead of STDIN")
	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	var input io.Reader = cli.in
	if *infile != "" {
		file, err := os.Open(*infile)
		if err != nil {
			return err
		}
		defer file.Close()
		input = file
	}

	response, err := cli.client.ImageLoad(input)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.JSON {
		return jsonmessage.DisplayJSONMessagesStream(response.Body, cli.out, cli.outFd, cli.isTerminalOut, nil)
	}

	_, err = io.Copy(cli.out, response.Body)
	return err
}
