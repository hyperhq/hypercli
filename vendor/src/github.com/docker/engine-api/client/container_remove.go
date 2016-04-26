package client

import (
	"encoding/json"
	"net/url"

	"github.com/docker/engine-api/types"
)

// ContainerRemove kills and removes a container from the docker host.
func (cli *Client) ContainerRemove(options types.ContainerRemoveOptions) ([]string, error) {
	var warnings []string
	query := url.Values{}
	if options.RemoveVolumes {
		query.Set("v", "1")
	}
	if options.RemoveLinks {
		query.Set("link", "1")
	}

	if options.Force {
		query.Set("force", "1")
	}

	resp, err := cli.delete("/containers/"+options.ContainerID, query, nil)
	if err == nil {
		json.NewDecoder(resp.body).Decode(&warnings)
	}
	ensureReaderClosed(resp)
	return warnings, err
}
