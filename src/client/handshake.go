package client

import (
	"errors"

	"github.com/ooclab/es"
	pjson "github.com/ooclab/otunnel/proto/json"
)

func handshake(conn es.Conn) error {
	jconn := pjson.NewConn(conn)

	if err := clientAuth(jconn); err != nil {
		return err
	}

	return nil
}

func clientAuth(c *pjson.Conn) error {
	resp, err := c.Request(map[string]interface{}{
		"action": "new",
	})
	if err != nil {
		return err
	}

	// fmt.Printf("got resp: %+v\n", resp)

	_, ok := resp["link_id"]
	// fmt.Println(linkID, ok)
	if !ok {
		return errors.New("can not find link_id in sayhi response")
	}

	// log.Debugf("got link %d", uint64(linkID.(float64)))
	// fmt.Println("linkID = ", linkID)

	return nil
}
