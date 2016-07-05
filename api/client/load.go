package client

import (
	"errors"
	"io"

	"golang.org/x/net/context"

	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/pkg/jsonmessage"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
)

// CmdLoad loads an image from a tar archive.
//
// The tar archive is read from STDIN by default, or from a tar archive file.
//
// Usage: docker load [OPTIONS]
func (cli *DockerCli) CmdLoad(args ...string) error {
	cmd := Cli.Subcmd("load", nil, Cli.DockerCommands["load"].Description, true)
	infile := cmd.String([]string{"i", "-input"}, "", "Read from a remote archive file compressed with gzip, bzip, or xz")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Do not show load process")
	cmd.Require(flag.Exact, 0)
	cmd.ParseFlags(args, true)

	if *infile == "" {
		return errors.New("remote archive must be specified via --input")
	}

	var input struct {
		FromSrc string `json:"fromSrc"`
		Quiet   bool   `json:"quiet"`
	}
	input.FromSrc = *infile
	input.Quiet = *quiet

	response, err := cli.client.ImageLoad(context.Background(), input)
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
