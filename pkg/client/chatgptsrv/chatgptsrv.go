// Code generated from jsonrpc schema by rpcgen v2.5.0; DO NOT EDIT.

package chatgptsrv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/vmkteam/appkit"
	"github.com/vmkteam/zenrpc/v2"
)

const name = "chatgptsrv"

var (
	// Always import time package. Generated models can contain time.Time fields.
	_ time.Time
)

type Client struct {
	rpcClient *rpcClient

	Chatgpt *svcChatgpt
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: time.Second * 30}
	}
	c := &Client{
		rpcClient: newRPCClient(endpoint, httpClient),
	}

	c.Chatgpt = newClientChatgpt(c.rpcClient)

	return c
}

type svcChatgpt struct {
	client *rpcClient
}

func newClientChatgpt(client *rpcClient) *svcChatgpt {
	return &svcChatgpt{
		client: client,
	}
}

func (c *svcChatgpt) Send(ctx context.Context, request string) (res string, err error) {
	_req := struct {
		Request string
	}{
		Request: request,
	}

	err = c.client.call(ctx, "chatgpt.Send", _req, &res)

	return
}

type rpcClient struct {
	endpoint string
	cl       *http.Client

	requestID uint64
}

func newRPCClient(endpoint string, httpClient *http.Client) *rpcClient {
	return &rpcClient{
		endpoint: endpoint,
		cl:       httpClient,
	}
}

func (rc *rpcClient) call(ctx context.Context, methodName string, request, result interface{}) error {
	// encode params
	bts, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode params: %w", err)
	}

	requestID := atomic.AddUint64(&rc.requestID, 1)
	requestIDBts := json.RawMessage(strconv.Itoa(int(requestID)))

	req := zenrpc.Request{
		Version: zenrpc.Version,
		ID:      &requestIDBts,
		Method:  methodName,
		Params:  bts,
	}

	ctx = appkit.NewCallerNameContext(ctx, name)

	res, err := rc.Exec(ctx, req)
	if err != nil {
		return err
	}

	if res == nil {
		return nil
	}

	if res.Error != nil {
		return res.Error
	}

	if res.Result == nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return json.Unmarshal(*res.Result, result)
}

// Exec makes http request to jsonrpc endpoint and returns json rpc response.
func (rc *rpcClient) Exec(ctx context.Context, rpcReq zenrpc.Request) (*zenrpc.Response, error) {
	if appkit.NotificationFromContext(ctx) {
		rpcReq.ID = nil
	}

	c, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("json marshal call failed: %w", err)
	}

	buf := bytes.NewReader(c)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rc.endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	appkit.SetXRequestIDFromCtx(ctx, req)

	// Do request
	resp, err := rc.cl.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, fmt.Errorf("make request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response (%d)", resp.StatusCode)
	}

	var zresp zenrpc.Response
	if rpcReq.ID == nil {
		return &zresp, nil
	}

	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body (%s) read failed: %w", bb, err)
	}

	if err = json.Unmarshal(bb, &zresp); err != nil {
		return nil, fmt.Errorf("json decode failed (%s): %w", bb, err)
	}

	return &zresp, nil
}
