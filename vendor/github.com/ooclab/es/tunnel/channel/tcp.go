package channel

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/ooclab/es"
	"github.com/ooclab/es/util"
	"github.com/sirupsen/logrus"

	tcommon "github.com/ooclab/es/tunnel/common"
)

type tcpChannel struct {
	// !IMPORTANT! atomic.AddInt64 in arm / x86_32
	// https://plus.ooclab.com/note/article/1285
	recv uint64
	send uint64

	tid      uint32
	cid      uint32
	outbound chan []byte
	conn     net.Conn

	closed         bool
	closedByRemote bool // FIXME!

	lock *sync.Mutex
}

func (c *tcpChannel) ID() uint32 {
	return c.cid
}

func (c *tcpChannel) String() string {
	return fmt.Sprintf(`[TCP Channel] %d-%d: L(%s), R(%s)`, c.tid, c.cid, c.conn.LocalAddr(), c.conn.RemoteAddr())
}

func (c *tcpChannel) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	closeConn(c.conn)

	logrus.Debugf("CLOSE tcp channel %s: recv = %d, send = %d", c, atomic.LoadUint64(&c.recv), atomic.LoadUint64(&c.send))
}

func (c *tcpChannel) IsClosedByRemote() bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.closedByRemote
}

func (c *tcpChannel) SetClosedByRemote() {
	c.lock.Lock()
	c.closedByRemote = true
	c.lock.Unlock()
}

func (c *tcpChannel) HandleIn(m *tcommon.TMSG) error {
	// TODO: 1. use write cached !
	// TODO: 2. use goroutine & channel to handle inbound message ?
	wLen, err := c.conn.Write(m.Payload)
	// FIXME: make sure write all data, BUT it seems that golang do it already!
	if wLen != len(m.Payload) {
		logrus.Errorf("tcp channel c.conn.Write error: wLen = %d != len(m.Payload) = %d", wLen, len(m.Payload))
	}
	if err != nil {
		logrus.Errorf("channel write failed: %s", err)
		return errors.New("write payload error")
	}

	atomic.AddUint64(&c.send, uint64(wLen))
	return nil
}

func (c *tcpChannel) Serve() error {
	// logrus.Debugf("start serve channel %s", c)

	// FIXME!
	defer func() {
		if r := recover(); r != nil {
			logrus.Warn("channel serve recovered: ", r)
		}
		if !c.closed {
			c.Close()
		}
	}()

	// link.outbound <- channel.conn.Read
	for {
		// IMPORTANT: buf read size is very important for speed!
		buf := make([]byte, 1024*16) // TODO: custom
		reqLen, err := c.conn.Read(buf)
		if err != nil {
			if c.closed || util.TCPisClosedConnError(err) {
				logrus.Debugf("channel %s is closed normally, quit read", c)
				return nil
			}
			if err != io.EOF {
				logrus.Warnf("channel %s recv failed: %s", c, err)
			}

			return err
		}

		m := &tcommon.TMSG{
			Type:      tcommon.MsgTypeChannelForward,
			TunnelID:  c.tid,
			ChannelID: c.cid,
			Payload:   buf[:reqLen],
		}
		// FIXME! panic: send on closed channel
		c.outbound <- append([]byte{es.LinkMsgTypeTunnel}, m.Bytes()...)
		atomic.AddUint64(&c.recv, uint64(reqLen))
	}
}
