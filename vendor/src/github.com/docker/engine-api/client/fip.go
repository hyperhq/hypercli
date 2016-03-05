package client

import (
	"encoding/json"
	"net/url"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

func (cli *Client) FipAllocate(count string) ([]string, error) {
	var result []string
	var v = url.Values{}
	v.Set("count", count)
	serverResp, err := cli.post("/fips/allocate?"+v.Encode(), nil, nil, nil)
	if err != nil {
		return result, err
	}

	json.NewDecoder(serverResp.body).Decode(&result)
	ensureReaderClosed(serverResp)
	return result, err
}

func (cli *Client) FipRelease(ip string) error {
	var v = url.Values{}
	v.Set("ip", ip)
	_, err := cli.post("/fips/release?"+v.Encode(), nil, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) FipAssociate(ip, container string) error {
	var v = url.Values{}
	v.Set("ip", ip)
	v.Set("container", container)
	_, err := cli.post("/fips/associate?"+v.Encode(), nil, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) FipDeassociate(container string) (string, error) {
	var result string
	var v = url.Values{}
	v.Set("container", container)
	resp, err := cli.post("/fips/deassociate?"+v.Encode(), nil, nil, nil)
	if err != nil {
		return "", err
	}
	json.NewDecoder(resp.body).Decode(&result)
	ensureReaderClosed(resp)
	return result, nil
}

func (cli *Client) FipList(options types.NetworkListOptions) ([]string, error) {
	query := url.Values{}
	if options.Filters.Len() > 0 {
		filterJSON, err := filters.ToParam(options.Filters)
		if err != nil {
			return nil, err
		}

		query.Set("filters", filterJSON)
	}
	var fips []string
	resp, err := cli.get("/fips", query, nil)
	if err != nil {
		return fips, err
	}
	err = json.NewDecoder(resp.body).Decode(&fips)
	ensureReaderClosed(resp)
	return fips, err
}
