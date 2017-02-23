package client

import (
	"os"
	"fmt"
	"net"
	"crypto/tls"
	"path"
	"bytes"
	"strconv"
	"encoding/json"
	"net/http/httputil"
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

func newFuncEndpointRequest(method, subpath string, query url.Values) (*http.Request, error) {
	endpoint := os.Getenv("HYPER_FUNC_ENDPOINT")
	if endpoint == "" {
		endpoint = "us-west-1.hyperfunc.io"
	}
	apiUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	apiUrl.Scheme = "https"
	apiUrl.Path = path.Join(apiUrl.Path, subpath)
	queryStr := query.Encode()
	if queryStr != "" {
		apiUrl.RawQuery = queryStr
	}
	req, err := http.NewRequest(method, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func funcEndpointRequestHijack(method, subpath string, query url.Values) (*net.Conn, error) {
	req, err := newFuncEndpointRequest(method, subpath, query)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	conn, err := tls.Dial("tcp", req.URL.Host + ":443", &tls.Config{})
	if err != nil {
		return nil, err
	}
	clientConn := httputil.NewClientConn(conn, nil)
	_, err = clientConn.Do(req)
	if err != nil {
		return nil, err
	}
	respConn, _ := clientConn.Hijack()
	return &respConn, nil
}

func funcEndpointRequest(method, subpath string, query url.Values) (*http.Response, error) {
	req, err := newFuncEndpointRequest(method, subpath, query)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	status := resp.StatusCode
	if status < 200 || status >= 400 {
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
	resp, err := funcEndpointRequest("POST", path.Join("call", name, fn.UUID), nil)
	defer ensureFuncRespClosed(resp)
	if err != nil {
		return nil, err
	}
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
	subpath := callId
	if wait {
		subpath += "/wait"
	}
	resp, err := funcEndpointRequest("GET", path.Join("output", name, fn.UUID, subpath), nil)
	defer ensureFuncRespClosed(resp)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cli *Client) FuncLogs(ctx context.Context, name, callId string, follow bool, tail string) (*net.Conn, error) {
	fn, _, err := cli.FuncInspectWithRaw(ctx, name)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	if callId != "" {
		query.Set("callid", callId)
	}
	if follow {
		query.Add("follow", strconv.FormatBool(follow))
	}
	if tail != "" {
		query.Add("tail", tail)
	}
	conn, err := funcEndpointRequestHijack("GET", path.Join("logs", name, fn.UUID, ""), query)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
