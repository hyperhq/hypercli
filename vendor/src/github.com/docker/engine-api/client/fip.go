package client

import (
	"encoding/json"
	"net/url"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

func (cli *Client) FipAllocate(ctx context.Context, count string) ([]string, error) {
	var result []string
	var v = url.Values{}
	v.Set("count", count)
	serverResp, err := cli.post(ctx, "/fips/allocate", v, nil, nil)
	if err != nil {
		return result, err
	}

	json.NewDecoder(serverResp.body).Decode(&result)
	ensureReaderClosed(serverResp)
	return result, err
}

func (cli *Client) FipRelease(ctx context.Context, ip string) error {
	var v = url.Values{}
	v.Set("ip", ip)
	_, err := cli.post(ctx, "/fips/release", v, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) FipAssociate(ctx context.Context, ip, container string) error {
	var v = url.Values{}
	v.Set("ip", ip)
	v.Set("container", container)
	_, err := cli.post(ctx, "/fips/associate", v, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) FipDisassociate(ctx context.Context, container string) (string, error) {
	var result string
	var v = url.Values{}
	v.Set("container", container)
	resp, err := cli.post(ctx, "/fips/deassociate", v, nil, nil)
	if err != nil {
		return "", err
	}
	json.NewDecoder(resp.body).Decode(&result)
	ensureReaderClosed(resp)
	return result, nil
}

func (cli *Client) FipList(ctx context.Context, options types.NetworkListOptions) ([]map[string]string, error) {
	query := url.Values{}
	if options.Filters.Len() > 0 {
		filterJSON, err := filters.ToParam(options.Filters)
		if err != nil {
			return nil, err
		}

		query.Set("filters", filterJSON)
	}
	var fips []map[string]string
	resp, err := cli.get(ctx, "/fips", query, nil)
	if err != nil {
		return fips, err
	}
	err = json.NewDecoder(resp.body).Decode(&fips)
	ensureReaderClosed(resp)
	return fips, err
}
