package channel

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es"

	tcommon "github.com/ooclab/es/tunnel/common"
)

type udpChannel struct {
	// !IMPORTANT! atomic.AddInt64 in arm / x86_32
	// https://plus.ooclab.com/note/article/1285
	recv uint64
	send uint64

	tid      uint32
	cid      uint32
	outbound chan []byte
	conn     net.Conn

	closed bool

	lock *sync.Mutex
}

func (c *udpChannel) ID() uint32 {
	return c.cid
}

func (c *udpChannel) String() string {
	return fmt.Sprintf(`[UDP Channel] %d-%d: L(%s), R(%s)`, c.tid, c.cid, c.conn.LocalAddr(), c.conn.RemoteAddr())
}

func (c *udpChannel) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	closeConn(c.conn)

	logrus.Debugf("CLOSE udp channel %s: recv = %d, send = %d", c, atomic.LoadUint64(&c.recv), atomic.LoadUint64(&c.send))
}

func (c *udpChannel) IsClosed() bool {
	return c.closed
}

func (c *udpChannel) HandleIn(m *tcommon.TMSG) error {
	// TODO: 1. use write cached !
	// TODO: 2. use goroutine & channel to handle inbound message ?
	wLen, err := c.conn.Write(m.Payload)
	// FIXME: make sure write all data, BUT it seems that golang do it already!
	if wLen != len(m.Payload) {
		logrus.Errorf("udp channel: c.conn.Write error: wLen = %d != len(m.Payload) = %d", wLen, len(m.Payload))
	}
	if err != nil {
		logrus.Errorf("channel write failed: %s", err)
		return errors.New("write payload error")
	}

	atomic.AddUint64(&c.send, uint64(wLen))
	return nil
}

func (c *udpChannel) Serve() error {
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
	// FIXME! udp should
	lastReceived := time.Now()
	for time.Since(lastReceived) < 6*time.Second {
		// IMPORTANT: buf read size is very important for speed!
		buf := make([]byte, 1024*16)                            // TODO: custom
		c.conn.SetReadDeadline(time.Now().Add(6 * time.Second)) // TODO: custom
		reqLen, err := c.conn.Read(buf)
		fmt.Println("--- reqLen, err = ", reqLen, err)
		if err != nil {
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

		// update the lastReceived
		lastReceived = time.Now()
	}
	return errors.New("timeout")
}
