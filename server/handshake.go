package server

import (
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es/emsg"
	pjson "github.com/ooclab/otunnel/proto/json"
)

func handshake(rawConn io.ReadWriteCloser) error {
	conn := emsg.NewConn(rawConn)
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

	logrus.Debugf("handle auth, go request: %+v\n", m)
	return c.Send(map[string]interface{}{
		"link_id": 123456,
		"hello":   "world",
	})
}
