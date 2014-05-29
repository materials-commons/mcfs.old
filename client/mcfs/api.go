package mcfs

import (
	"fmt"
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/client/util"
	"github.com/materials-commons/mcfs/protocol"
	"net"
)

// NewClient creates a new connection to the file server.
func NewClient(host string, port int) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	m := util.NewGobMarshaler(conn)
	c := &Client{
		MarshalUnmarshaler: m,
		conn:               conn,
	}
	return c, nil
}

// Close closes the connection to the server.
func (c *Client) Close() {
	c.conn.Close()
}

// Login performs a login request.
func (c *Client) Login(user, apikey string) error {
	req := protocol.LoginReq{
		User:   user,
		APIKey: apikey,
	}

	_, err := c.doRequest(req)
	return err
}

// Logout performs a logout request.
func (c *Client) Logout() error {
	req := protocol.LogoutReq{}
	_, err := c.doRequest(req)
	return err
}

// doRequest executes a request and waits for the response.
func (c *Client) doRequest(arg interface{}) (interface{}, error) {
	req := &protocol.Request{
		Req: arg,
	}

	if err := c.Marshal(req); err != nil {
		return nil, err
	}

	var resp protocol.Response

	if err := c.Unmarshal(&resp); err != nil {
		return nil, err
	}

	if resp.Status != mc.ErrorCodeSuccess {
		return resp.Resp, mc.ErrorCodeToError(resp.Status)
	}

	return resp.Resp, nil
}

// doRequestNoResp executes a request that doesn't expect a response
func (c *Client) doRequestNoResp(arg interface{}) error {
	req := &protocol.Request{Req: arg}
	if err := c.Marshal(req); err != nil {
		return err
	}
	return nil
}
