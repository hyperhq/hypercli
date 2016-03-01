package client

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// SnapshotList returns the snapshots configured in the docker host.
func (cli *Client) SnapshotList(filter filters.Args) (types.SnapshotsListResponse, error) {
	var snapshots types.SnapshotsListResponse
	query := url.Values{}

	if filter.Len() > 0 {
		filterJSON, err := filters.ToParam(filter)
		if err != nil {
			return snapshots, err
		}
		query.Set("filters", filterJSON)
	}
	resp, err := cli.get("/snapshots", query, nil)
	if err != nil {
		return snapshots, err
	}

	err = json.NewDecoder(resp.body).Decode(&snapshots)
	ensureReaderClosed(resp)
	return snapshots, err
}

// SnapshotInspect returns the information about a specific snapshot in the docker host.
func (cli *Client) SnapshotInspect(snapshotID string) (types.Snapshot, error) {
	var snapshot types.Snapshot
	resp, err := cli.get("/snapshots/"+snapshotID, nil, nil)
	if err != nil {
		if resp.statusCode == http.StatusNotFound {
			return snapshot, snapshotNotFoundError{snapshotID}
		}
		return snapshot, err
	}
	err = json.NewDecoder(resp.body).Decode(&snapshot)
	ensureReaderClosed(resp)
	return snapshot, err
}

// SnapshotCreate creates a snapshot in the docker host.
func (cli *Client) SnapshotCreate(options types.SnapshotCreateRequest) (types.Snapshot, error) {
	var snapshot types.Snapshot
	resp, err := cli.post("/snapshots/create", nil, options, nil)
	if err != nil {
		return snapshot, err
	}
	err = json.NewDecoder(resp.body).Decode(&snapshot)
	ensureReaderClosed(resp)
	return snapshot, err
}

// SnapshotRemove removes a snapshot from the docker host.
func (cli *Client) SnapshotRemove(snapshotID string) error {
	resp, err := cli.delete("/snapshots/"+snapshotID, nil, nil)
	ensureReaderClosed(resp)
	return err
}
