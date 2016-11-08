package emsg

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/ooclab/es/ecrypt"
)

type Conn struct {
	conn   net.Conn
	cipher *ecrypt.Cipher
}

func NewConn(conn net.Conn) *Conn {
	c := &Conn{
		conn: conn,
	}
	return c
}

func (c *Conn) SetCipher(cipher *ecrypt.Cipher) {
	c.cipher = cipher
}

// 读取一条消息
func (c *Conn) Recv() (*EMSG, error) {
	// TODO: 优雅地读取消息
	// - 增加消息体之间的分割符
	// - 错误读取恢复/丢弃
	head, err := c.mustRecv(4)
	if err != nil {
		return nil, err
	}

	var dlen uint32
	dlen = binary.LittleEndian.Uint32(head)

	data, err := c.mustRecv(dlen)
	if err != nil {
		return nil, err
	}
	if c.cipher != nil {
		// 解密消息
		c.cipher.Decrypt(data[0:dlen], data[0:dlen])
	}

	// 创建 EMSG
	return LoadEMSG(data)
}

func (c *Conn) Send(m *EMSG) error {
	b := m.Bytes()

	if c.cipher != nil {
		c.cipher.Encrypt(b, b)
	}

	buf := new(bytes.Buffer)

	dlen := uint32(len(b))
	err := binary.Write(buf, binary.LittleEndian, dlen)
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, b)
	if err != nil {
		return err
	}

	// TODO: make sure write exactly data
	_, err = c.conn.Write(buf.Bytes())
	return err
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

// 从 net.Conn 读取指定长度的字节流
func (c *Conn) mustRecv(dlen uint32) ([]byte, error) {
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
