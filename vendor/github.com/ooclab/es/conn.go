package es

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/ooclab/es/ecrypt"
)

const (
	maxMessageLength = 1024*64 - 1
)

// common error define
var (
	ErrBufferIsShort  = errors.New("buffer is short")
	ErrMaxLengthLimit = errors.New("max length limit")
)

// Conn is a interface a Conn
type Conn interface {
	Recv() (message []byte, err error)
	Send(message []byte) error
	Close() error
}

// BaseConn is the basic connection type
type BaseConn struct {
	conn io.ReadWriteCloser
}

// NewBaseConn create a base Conn object
func NewBaseConn(conn io.ReadWriteCloser) Conn {
	return &BaseConn{
		conn: conn,
	}
}

// Recv read a message from this Conn
func (c *BaseConn) Recv() (message []byte, err error) {
	head, err := c.mustRecv(2)
	if err != nil {
		return
	}
	return c.mustRecv(binary.BigEndian.Uint16(head))
}

// Send send a message to this Conn
func (c *BaseConn) Send(message []byte) error {
	dlen := uint16(len(message))
	if dlen > maxMessageLength {
		return ErrMaxLengthLimit
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, dlen)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, message)
	if err != nil {
		return err
	}

	// TODO: make sure write exactly data
	_, err = c.conn.Write(buf.Bytes())
	return err
}

// Close close a Conn
func (c *BaseConn) Close() error {
	return c.conn.Close()
}

func (c *BaseConn) mustRecv(dlen uint16) ([]byte, error) {
	data := make([]byte, dlen)
	for i := 0; i < int(dlen); {
		n, err := c.conn.Read(data[i:])
		if err != nil {
			return nil, err
		}
		i += n
	}
	return data, nil
}

// SafeConn ecrypt Conn
type SafeConn struct {
	BaseConn
	cipher *ecrypt.Cipher
}

// NewSafeConn create a safe Conn
func NewSafeConn(conn io.ReadWriteCloser, cipher *ecrypt.Cipher) Conn {
	c := &SafeConn{
		cipher: cipher,
	}
	c.conn = conn
	return c
}

// Recv read a message from this Conn
func (c *SafeConn) Recv() (message []byte, err error) {
	head, err := c.mustRecv(2)
	if err != nil {
		return
	}
	c.cipher.Decrypt(head[0:2], head[0:2])
	message, err = c.mustRecv(binary.BigEndian.Uint16(head))
	if err == nil {
		c.cipher.Decrypt(message, message)
	}
	return
}

// Send send a message to this Conn
func (c *SafeConn) Send(message []byte) error {
	dlen := uint16(len(message))
	if dlen > maxMessageLength {
		return ErrMaxLengthLimit
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, dlen)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.BigEndian, message)
	if err != nil {
		return err
	}

	// TODO: make sure write exactly data
	b := buf.Bytes()
	c.cipher.Encrypt(b, b)
	_, err = c.conn.Write(b)
	return err
}

// TODO: LargeMessageConn for send large message
