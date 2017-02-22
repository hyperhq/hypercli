package client

import (
	"os"
	"fmt"
	"path"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

func ensureFuncRespClosed(resp *http.Response) {
	if resp != nil {
		resp.Body.Close()
	}
}

func (cli *Client) FuncEndpointRequest(method, resource, name, uuid, query string) (*http.Response, error) {
	endpoint := os.Getenv("HYPER_FUNC_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://us-west-1.hyperfunc.io/"
	}
	api, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	api.Path = path.Join(api.Path, resource, name, uuid, query)

	httpClient := &http.Client{}
	req, err := http.NewRequest(method, api.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	// defer resp.Body.Close()

	status := resp.StatusCode
	if status != 101 && (status < 200 || status >= 400) {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		return resp, fmt.Errorf("Error response from server: %s", bytes.TrimSpace(body))
	}

	return resp, nil
}

func (cli *Client) FuncCreate(ctx context.Context, opts types.Func) (types.Func, error) {
	var fn types.Func
	resp, err := cli.post(ctx, "/funcs/create", nil, opts, nil)
	if err != nil {
		return fn, err
	}
	err = json.NewDecoder(resp.body).Decode(&fn)
	ensureReaderClosed(resp)
	return fn, err
}

func (cli *Client) FuncUpdate(ctx context.Context, name string, opts types.Func) (types.Func, error) {
	var fn types.Func
	resp, err := cli.put(ctx, "/funcs/"+name, nil, opts, nil)
	if err != nil {
		return fn, err
	}
	err = json.NewDecoder(resp.body).Decode(&fn)
	ensureReaderClosed(resp)
	return fn, err
}

func (cli *Client) FuncDelete(ctx context.Context, name string) error {
	resp, err := cli.delete(ctx, "/funcs/"+name, nil, nil)
	ensureReaderClosed(resp)
	return err
}

func (cli *Client) FuncList(ctx context.Context, opts types.FuncListOptions) ([]types.Func, error) {
	var fns = []types.Func{}
	query := url.Values{}

	if opts.Filters.Len() > 0 {
		filterJSON, err := filters.ToParamWithVersion(cli.version, opts.Filters)
		if err != nil {
			return fns, err
		}
		query.Set("filters", filterJSON)
	}
	resp, err := cli.get(ctx, "/funcs", query, nil)
	if err != nil {
		return fns, err
	}

	err = json.NewDecoder(resp.body).Decode(&fns)
	ensureReaderClosed(resp)
	return fns, err
}

func (cli *Client) FuncInspect(ctx context.Context, name string) (types.Func, error) {
	fn, _, err := cli.FuncInspectWithRaw(ctx, name)
	return fn, err
}

func (cli *Client) FuncInspectWithRaw(ctx context.Context, name string) (types.Func, []byte, error) {
	var fn types.Func
	resp, err := cli.get(ctx, "/funcs/"+name, nil, nil)
	if err != nil {
		if resp.statusCode == http.StatusNotFound {
			return fn, nil, funcNotFoundError{name}
		}
		return fn, nil, err
	}
	defer ensureReaderClosed(resp)

	body, err := ioutil.ReadAll(resp.body)
	if err != nil {
		return fn, nil, err
	}
	rdr := bytes.NewReader(body)
	err = json.NewDecoder(rdr).Decode(&fn)
	return fn, body, err
}

func (cli *Client) FuncCall(ctx context.Context, name string) (*types.FuncCallResponse, error) {
	fn, _, err := cli.FuncInspectWithRaw(ctx, name)
	if err != nil {
		return nil, err
	}
	resp, err := cli.FuncEndpointRequest("POST", "call", name, fn.UUID, "")
	if err != nil {
		return nil, err
	}
	defer ensureFuncRespClosed(resp)
	var ret types.FuncCallResponse
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (cli *Client) FuncGet(ctx context.Context, name, callId string, wait bool) ([]byte, error) {
	fn, _, err := cli.FuncInspectWithRaw(ctx, name)
	if err != nil {
		return nil, err
	}
	query := callId
	if wait {
		query += "/wait"
	}
	resp, err := cli.FuncEndpointRequest("GET", "output", name, fn.UUID, query)
	if err != nil {
		return nil, err
	}
	defer ensureFuncRespClosed(resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
