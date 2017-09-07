package tunnel

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es"
	"github.com/ooclab/es/tunnel/channel"
	tcommon "github.com/ooclab/es/tunnel/common"
	"github.com/ooclab/es/util"
)

type TunnelConfig struct {
	ID         uint32
	Proto      string // TCP or UDP
	LocalHost  string
	LocalPort  int
	RemoteHost string
	RemotePort int
	Reverse    bool
}

func (c *TunnelConfig) RemoteConfig() *TunnelConfig {
	return &TunnelConfig{
		Proto:      c.Proto,
		LocalHost:  c.RemoteHost,
		LocalPort:  c.RemotePort,
		RemoteHost: c.LocalHost,
		RemotePort: c.LocalPort,
		Reverse:    !c.Reverse,
	}
}

func (c *TunnelConfig) String() string {
	if c.Reverse {
		return fmt.Sprintf("%s: L(%s:%d) <- R(%s:%d)", c.Proto, c.LocalHost, c.LocalPort, c.RemoteHost, c.RemotePort)
	} else {
		return fmt.Sprintf("%s: L(%s:%d) -> R(%s:%d)", c.Proto, c.LocalHost, c.LocalPort, c.RemoteHost, c.RemotePort)
	}
}

// Tunnel define a tunnel struct
type Tunnel struct {
	ID          uint32
	Config      *TunnelConfig
	cpool       *channel.Pool
	outbound    chan []byte
	manager     *Manager
	openChannel func(*tcommon.TMSG) (channel.Channel, error)
	listenFunc  func() error
}

func newTunnel(manager *Manager, cfg *TunnelConfig) *Tunnel {
	t := &Tunnel{
		ID:       cfg.ID,
		Config:   cfg,
		cpool:    channel.NewPool(),
		outbound: manager.outbound,
		manager:  manager,
	}
	switch cfg.Proto {
	case "tcp":
		t.openChannel = t.openTCPChannel
		t.listenFunc = t.listenTCP
	case "udp":
		t.openChannel = t.openUDPChannel
		t.listenFunc = t.listenUDP
	default:
		logrus.Errorf("can not be here!")
		return nil
	}
	return t
}

func (t *Tunnel) String() string {
	cfg := t.Config
	if cfg.Reverse {
		return fmt.Sprintf("%d L:%s:%d <- R:%s:%d", t.ID, cfg.LocalHost, cfg.LocalPort, cfg.RemoteHost, cfg.RemotePort)
	}
	return fmt.Sprintf("%d L:%s:%d -> R:%s:%d", t.ID, cfg.LocalHost, cfg.LocalPort, cfg.RemoteHost, cfg.RemotePort)
}

func (t *Tunnel) HandleIn(m *tcommon.TMSG) (err error) {
	c := t.cpool.Get(m.ChannelID)
	if t.Config.Reverse {
		if c == nil {
			c, err = t.openChannel(m)
			if err != nil {
				return err
			}
			logrus.Debugf("HandleIn: OPEN tcp channel %s success", c)
		}
	} else {
		// forward tunnel
		if c == nil {
			logrus.Errorf("can not find channel %d:%d", m.TunnelID, m.ChannelID)
			return errors.New("no such channel")
		}
	}

	return c.HandleIn(m)
}

func (t *Tunnel) openTCPChannel(m *tcommon.TMSG) (channel.Channel, error) {
	// (reverse tunnel) need to setup a connect to localhost:localport
	cfg := t.Config
	addrS := fmt.Sprintf("%s:%d", cfg.LocalHost, cfg.LocalPort)
	addr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		logrus.Warnf("resolve %s failed: %s", addrS, err)
		// TODO: notice remote endpoint ?
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logrus.Errorf("dial %s failed: %s", addrS, err.Error())
		// TODO: try again ?
		return nil, err
	}

	// IMPORTANT! create channel by ID!
	c := t.cpool.NewByID(m.ChannelID, t.ID, t.outbound, conn)
	go t.ServeChannel(c)
	return c, nil
}

func (t *Tunnel) openUDPChannel(m *tcommon.TMSG) (channel.Channel, error) {
	// (reverse tunnel) need to setup a connect to localhost:localport
	cfg := t.Config
	addrS := fmt.Sprintf("%s:%d", cfg.LocalHost, cfg.LocalPort)
	addr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		logrus.Warnf("resolve %s failed: %s", addrS, err)
		// TODO: notice remote endpoint ?
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logrus.Errorf("dial %s failed: %s", addrS, err.Error())
		// TODO: try again ?
		return nil, err
	}

	// IMPORTANT! create channel by ID!
	c := t.cpool.NewByID(m.ChannelID, t.ID, t.outbound, conn)
	go t.ServeChannel(c)
	return c, nil
}

func (t *Tunnel) NewChannelByConn(conn net.Conn) channel.Channel {
	if t.Config.Reverse {
		logrus.Errorf("reverse tunnel can not create channel use random ID!")
		return nil
	}
	return t.cpool.New(t.ID, t.outbound, conn)
}

func (t *Tunnel) ServeChannel(c channel.Channel) {
	if err := c.Serve(); err != nil {
		if !c.IsClosedByRemote() {
			t.closeRemoteChannel(c.ID())
		}
	}
	if t.cpool.Exist(c.ID()) {
		t.cpool.Delete(c)
	}
}

func (t *Tunnel) closeRemoteChannel(cid uint32) {
	// FIXME! temp fix "panic: send on closed channel"
	defer func() {
		if r := recover(); r != nil {
			logrus.Error("t.closeRemoteChannel recovered: ", r)
		}
	}()
	logrus.Debugf("prepare notice remote endpoint to close channel %d", cid)
	m := &tcommon.TMSG{
		Type:      tcommon.MsgTypeChannelClose,
		TunnelID:  t.ID,
		ChannelID: cid,
	}
	// FIXME! panic: send on closed channel
	t.outbound <- append([]byte{es.LinkMsgTypeTunnel}, m.Bytes()...)
	logrus.Debugf("notice remote endpoint to close channel %d done", cid)
}

func (t *Tunnel) HandleChannelClose(m *tcommon.TMSG) error {
	c := t.cpool.Get(m.ChannelID)
	if c == nil {
		logrus.Warnf("can not find channel %d:%d", m.TunnelID, m.ChannelID)
		return errors.New("no such channel")
	}

	// TODO: more clean!
	c.Close()
	c.SetClosedByRemote()
	t.cpool.Delete(c)
	return nil
}

func (t *Tunnel) Listen() error {
	if t.Config.Reverse {
		// reverse tunnel can not listen
		return nil
	}
	return t.listenFunc()
}

func (t *Tunnel) listenTCP() error {
	host, port := t.Config.LocalHost, t.Config.LocalPort
	key := t.manager.lpool.TCPKey(host, port)

	if t.manager.lpool.Exist(key) {
		// the listen address is exist in lpool already
		logrus.Errorf("start listen for %s:%d failed, it's existed already.", host, port)
		return errors.New("listen address is existed")
	}

	// start listen
	addr := fmt.Sprintf("%s:%d", host, port)
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if nil != err {
		logrus.Fatalln(err)
	}
	l, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		// the listen address is taken by another program
		logrus.Errorf("start listen on %s failed: %s", addr, err)
		return err
	}

	// save listen
	t.manager.lpool.Add(key, newTCPListenTarget(t, host, port, l))

	go func() {
		// defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				if util.TCPisClosedConnError(err) {
					logrus.Debugf("the listener of %s is closed", t)
				} else {
					logrus.Errorf("accept new client failed: %s", err)
				}
				break
			}
			logrus.Debugf("tunnel %s accept new client %s", t.String(), conn.RemoteAddr())

			c := t.NewChannelByConn(conn)
			go t.ServeChannel(c)
			logrus.Debugf("listenTCP: OPEN channel %s success", c)
		}
	}()

	logrus.Debugf("start listen tunnel %s success", t)
	return nil
}

func (t *Tunnel) listenUDP() error {
	host, port := t.Config.LocalHost, t.Config.LocalPort
	key := t.manager.lpool.UDPKey(host, port)

	if t.manager.lpool.Exist(key) {
		// the listen address is exist in lpool already
		logrus.Errorf("start udp listen for %s:%d failed, it's existed already.", host, port)
		return errors.New("listen udp address is existed")
	}

	// start listen
	addr := fmt.Sprintf("%s:%d", host, port)
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if nil != err {
		logrus.Fatalln(err)
	}
	conn, err := net.ListenUDP("tcp", laddr)
	if err != nil {
		// the listen address is taken by another program
		logrus.Errorf("start listen on %s failed: %s", addr, err)
		return err
	}

	// save listen
	t.manager.lpool.Add(key, newUDPListenTarget(t, host, port, conn))

	go func() {
		buf := make([]byte, 1400) // FIXME!
		for {
			n, raddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				// FIXME!
				if strings.Contains(err.Error(), "use of closed network connection") {
					logrus.Debugf("conn is closed, quit recv()")
					return
				}
				logrus.Errorf("ReadFromUDP error: %s", err)
				return
			}
			fmt.Println("--- n, raddr, err = ", n, raddr, err)
		}
	}()

	logrus.Debugf("start listen tunnel %s success", t)
	return nil
}
