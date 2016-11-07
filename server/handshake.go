package server

import (
	"github.com/ooclab/es/emsg"
	pjson "github.com/ooclab/otunnel/proto/json"
)

func handshake(conn *emsg.Conn) error {
	jconn := pjson.NewConn(conn)

	if err := handleAuth(jconn); err != nil {
		return err
	}

	return nil
}

func handleAuth(c *pjson.Conn) error {
	_, err := c.Recv()
	if err != nil {
		return err
	}

	// username := req["username"]
	// password := req["password"]
	//
	// if username != "admin" {
	// 	return errors.New("not admin user")
	// }

	// log.Debugf("req = %+v\n", req)
	return c.Send(map[string]interface{}{
		"link_id": 123456,
		"hello":   "world",
	})
}
