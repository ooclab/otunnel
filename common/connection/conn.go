package connection

import (
	"net"

	"github.com/ooclab/es/ecrypt"
)

type Conn struct {
	conn   net.Conn
	cipher *ecrypt.Cipher
}

func NewConn(conn net.Conn, cipher *ecrypt.Cipher) *Conn {
	return &Conn{
		conn:   conn,
		cipher: cipher,
	}
}

func (c *Conn) Read(b []byte) (int, error) {
	n, err := c.conn.Read(b)
	if err == nil {
		c.cipher.Decrypt(b, b)
	}
	return n, err
}

func (c *Conn) Write(b []byte) (int, error) {
	c.cipher.Encrypt(b, b)
	return c.conn.Write(b)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) CloseRead() {
	if conn, ok := c.conn.(*net.TCPConn); ok {
		conn.CloseRead()
	}
}

func (c *Conn) CloseWrite() {
	if conn, ok := c.conn.(*net.TCPConn); ok {
		conn.CloseWrite()
	}
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
