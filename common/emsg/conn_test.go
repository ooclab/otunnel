package emsg

import (
	"net"
	"ooclab/emsg/serial"
	"testing"
)

func Test_Conn(t *testing.T) {

	cipher := NewCipher("aes256cfb", []byte("RmHUgueMklsBpPt53cLoxJ6AjiYCDrz4f7hFOK0Z"))
	done := make(chan bool)

	payload, err := serial.JSONDumps(map[string]int{
		"k1": 1,
		"k2": 2,
		"k3": 3,
	})
	if err != nil {
		t.Error("gen payload failed:", err)
	}

	m := &EMSG{
		Version: 1,
		Type:    2,
		Proto:   3,
		Id:      654321,
		Payload: payload,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:2001")
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		conn := NewConn(c)
		conn.SetCipher(cipher)
		defer conn.Close()

		in_msg, err := conn.Recv()
		if err != nil {
			t.Error(err)
		}

		if in_msg.Version != m.Version ||
			in_msg.Id != m.Id ||
			in_msg.Type != m.Type ||
			in_msg.Proto != m.Proto ||
			string(in_msg.Payload) != string(m.Payload) {
			t.Error("msg mismatch!")
		}

		close(done)
	}()

	c, err := net.Dial("tcp", "127.0.0.1:2001")
	if err != nil {
		t.Error(err)
	}
	conn := NewConn(c)
	conn.SetCipher(cipher)
	defer conn.Close()

	err = conn.Send(m)
	if err != nil {
		t.Error(err)
	}

	<-done
}
