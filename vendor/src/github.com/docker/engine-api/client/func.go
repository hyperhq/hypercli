package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
)

func newFuncEndpointRequest(method, subpath string, query url.Values, body io.Reader) (*http.Request, error) {
	endpoint := os.Getenv("HYPER_FUNC_ENDPOINT")
	if endpoint == "" {
		endpoint = "us-west-1.hyperfunc.io"
	}
	apiURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	apiURL.Scheme = "https"
	apiURL.Path = path.Join(apiURL.Path, subpath)
	queryStr := query.Encode()
	if queryStr != "" {
		apiURL.RawQuery = queryStr
	}
	req, err := http.NewRequest(method, apiURL.String(), body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func funcEndpointRequestHijack(req *http.Request) (net.Conn, error) {
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")
	conn, err := tls.Dial("tcp", req.URL.Host+":443", &tls.Config{})
	if err != nil {
		return nil, err
	}
	clientConn := httputil.NewClientConn(conn, nil)
	resp, err := clientConn.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 101 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Error response from server: %s", bytes.TrimSpace(body))
	}
	respConn, _ := clientConn.Hijack()
	return respConn, nil
}

func funcEndpointRequest(req *http.Request) (*http.Response, error) {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	status := resp.StatusCode
	if status < 200 || status >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Error response from server: %s", bytes.TrimSpace(body))
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

func (cli *Client) FuncCall(ctx context.Context, name string, stdin io.Reader) (*types.FuncCallResponse, error) {
	fn, _, err := cli.FuncInspectWithRaw(ctx, name)
	if err != nil {
		return nil, err
	}
	req, err := newFuncEndpointRequest("POST", path.Join("call", name, fn.UUID), nil, stdin)
	if err != nil {
		return nil, err
	}
	resp, err := funcEndpointRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
	req, err := newFuncEndpointRequest("GET", path.Join("output", name, fn.UUID, subpath), nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := funcEndpointRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cli *Client) FuncLogs(ctx context.Context, name, callId string, follow bool, tail string) (io.ReadCloser, error) {
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
	req, err := newFuncEndpointRequest("GET", path.Join("logs", name, fn.UUID, ""), query, nil)
	if err != nil {
		return nil, err
	}
	if follow {
		conn, err := funcEndpointRequestHijack(req)
		if err != nil {
			return nil, err
		}
		return conn.(io.ReadCloser), nil
	}
	resp, err := funcEndpointRequest(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
