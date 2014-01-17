package mcfs

import (
	"fmt"
	"github.com/materials-commons/materials/util"
	"github.com/materials-commons/mcfs/protocol"
	"net"
)

func NewClient(host string, port uint) (*Client, error) {
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

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) Login(user, apikey string) error {
	req := protocol.LoginReq{
		User:   user,
		ApiKey: apikey,
	}

	_, err := c.doRequest(req)
	return err
}

func (c *Client) Logout() error {
	req := protocol.LogoutReq{}
	_, err := c.doRequest(req)
	return err
}

func (c *Client) CreateProject(projectName string) (*Project, error) {
	req := &protocol.CreateProjectReq{
		Name: projectName,
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	switch t := resp.(type) {
	case protocol.CreateProjectResp:
		return &Project{t.ProjectID, t.DataDirID}, nil
	default:
		fmt.Printf("1 %s %T\n", ErrBadResponseType, t)
		return nil, ErrBadResponseType
	}
}

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

	if resp.Type != protocol.ROk {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp.Resp, nil
}
