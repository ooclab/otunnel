package server

import (
	"github.com/ooclab/es"
	pjson "github.com/ooclab/otunnel/pkg/proto/json"
	"github.com/sirupsen/logrus"
)

func handshake(conn es.Conn) error {
	jconn := pjson.NewConn(conn)

	if err := handleAuth(jconn); err != nil {
		return err
	}

	return nil
}

func handleAuth(c *pjson.Conn) error {
	m, err := c.Recv()
	if err != nil {
		return err
	}

	// username := req["username"]
	// password := req["password"]
	//
	// if username != "admin" {
	// 	return errors.New("not admin user")
	// }

	logrus.Debugf("handle auth, got request: %+v", m)
	return c.Send(map[string]interface{}{
		"link_id": 123456,
		"hello":   "world",
	})
}
