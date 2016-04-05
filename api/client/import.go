package client

import (
	"fmt"
	"io"
	"os"

	Cli "github.com/hyperhq/hypercli/cli"
	"github.com/hyperhq/hypercli/opts"
	"github.com/hyperhq/hypercli/pkg/jsonmessage"
	flag "github.com/hyperhq/hypercli/pkg/mflag"
	"github.com/hyperhq/hypercli/pkg/urlutil"
	"github.com/hyperhq/hypercli/reference"
	"github.com/docker/engine-api/types"
)

// CmdImport creates an empty filesystem image, imports the contents of the tarball into the image, and optionally tags the image.
//
// The URL argument is the address of a tarball (.tar, .tar.gz, .tgz, .bzip, .tar.xz, .txz) file or a path to local file relative to docker client. If the URL is '-', then the tar file is read from STDIN.
//
// Usage: docker import [OPTIONS] file|URL|- [REPOSITORY[:TAG]]
func (cli *DockerCli) Import(args ...string) error {
	cmd := Cli.Subcmd("import", []string{"file|URL|- [REPOSITORY[:TAG]]"}, Cli.DockerCommands["import"].Description, true)
	flChanges := opts.NewListOpts(nil)
	cmd.Var(&flChanges, []string{"c", "-change"}, "Apply Dockerfile instruction to the created image")
	message := cmd.String([]string{"m", "-message"}, "", "Set commit message for imported image")
	cmd.Require(flag.Min, 1)

	cmd.ParseFlags(args, true)

	var (
		in         io.Reader
		tag        string
		src        = cmd.Arg(0)
		srcName    = src
		repository = cmd.Arg(1)
		changes    = flChanges.GetAll()
	)

	if cmd.NArg() == 3 {
		fmt.Fprintf(cli.err, "[DEPRECATED] The format 'file|URL|- [REPOSITORY [TAG]]' has been deprecated. Please use file|URL|- [REPOSITORY[:TAG]]\n")
		tag = cmd.Arg(2)
	}

	if repository != "" {
		//Check if the given image name can be resolved
		if _, err := reference.ParseNamed(repository); err != nil {
			return err
		}
	}

	if src == "-" {
		in = cli.in
	} else if !urlutil.IsURL(src) {
		srcName = "-"
		file, err := os.Open(src)
		if err != nil {
			return err
		}
		defer file.Close()
		in = file
	}

	options := types.ImageImportOptions{
		Source:         in,
		SourceName:     srcName,
		RepositoryName: repository,
		Message:        *message,
		Tag:            tag,
		Changes:        changes,
	}

	responseBody, err := cli.client.ImageImport(options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.out, cli.outFd, cli.isTerminalOut, nil)
}
