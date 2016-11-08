package json

import (
	"encoding/json"

	"github.com/ooclab/otunnel/common/emsg"
)

// Conn is a json proto connection
type Conn struct {
	conn *emsg.Conn
}

// NewConn create a server side connection object
func NewConn(conn *emsg.Conn) *Conn {
	return &Conn{
		conn: conn,
	}
}

// Request send and recv payload
func (c *Conn) Request(payload map[string]interface{}) (map[string]interface{}, error) {
	if err := c.Send(payload); err != nil {
		return nil, err
	}
	return c.Recv()
}

// Recv recv a map[string]interface{} type payload
func (c *Conn) Recv() (map[string]interface{}, error) {
	// 等待回复
	msg, err := c.conn.Recv()
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// Send send a map[string]interface{} type payload
func (c *Conn) Send(payload map[string]interface{}) error {
	_payload, _ := json.Marshal(payload)

	m := &emsg.EMSG{
		Type:    emsg.MSG_TYPE_REQ,
		Payload: _payload,
	}
	return c.conn.Send(m)
}

// Close close connection
func (c *Conn) Close() {
	c.conn.Close()
}
