package server

import (
	"crypto/sha1"
	"crypto/tls"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/pbkdf2"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es/ecrypt"
	"github.com/ooclab/es/link"
	"github.com/ooclab/otunnel/common/connection"
	"github.com/ooclab/otunnel/common/emsg"
	"github.com/urfave/cli"
	"github.com/xtaci/kcp-go"

	"net/http"
	_ "net/http/pprof"
)

type KCPServerConfig struct {
	Key          string `json:"key"`
	Crypt        string `json:"crypt"`
	Mode         string `json:"mode"`
	MTU          int    `json:"mtu"`
	SndWnd       int    `json:"sndwnd"`
	RcvWnd       int    `json:"rcvwnd"`
	DataShard    int    `json:"datashard"`
	ParityShard  int    `json:"parityshard"`
	DSCP         int    `json:"dscp"`
	NoComp       bool   `json:"nocomp"`
	AckNodelay   bool   `json:"acknodelay"`
	NoDelay      int    `json:"nodelay"`
	Interval     int    `json:"interval"`
	Resend       int    `json:"resend"`
	NoCongestion int    `json:"nc"`
	SockBuf      int    `json:"sockbuf"`
	KeepAlive    int    `json:"keepalive"`
	Log          string `json:"log"`
}

var (
	// SALT is use for pbkdf2 key expansion
	KCP_SALT = "kcp-go"
)

// StartDefaultListener run a default listener
func StartDefaultListener(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

// StartTLSListener run a tls listener
func StartTLSListener(addr string, caFile string, certFile string, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logrus.Errorf("load X509KeyPair failed: %s", err)
		return nil, err
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", addr, &config)
}

// StartAESListener run a aes listener
func StartAESListener(addr string, secret string) (net.Listener, error) {
	// TODO: does it needed use secret string here?
	return StartDefaultListener(addr)
}

// Server define a server object
type Server struct {
	Proto string
	Type  string

	addr string

	// aes connection needed!
	secret string

	// tls connection needed!
	caFile   string
	keyFile  string
	certFile string
}

// NewServer create a server object
func NewServer(c *cli.Context) *Server {
	pprof := c.String("pprof")
	if pprof != "" {
		go func() {
			logrus.Println(http.ListenAndServe(pprof, nil))
		}()
	}

	addr := c.Args().First()
	if len(addr) == 0 {
		addr = ":10000"
	}

	s := &Server{
		Proto:    c.String("proto"),
		addr:     addr,
		secret:   c.String("secret"),
		caFile:   c.String("ca"),
		certFile: c.String("cert"),
		keyFile:  c.String("key"),
	}

	if len(s.secret) > 0 {
		s.Type = "aes"
	} else if len(s.certFile) > 0 && len(s.keyFile) > 0 {
		s.Type = "tls"
	} else {
		s.Type = "default"
	}

	return s
}

// Start run a server
func (s *Server) Start() {
	switch s.Proto {
	case "tcp":
		s.startTCP()
	case "kcp":
		s.startKCP()
	default:
		logrus.Errorf("unknown link proto: %s", s.Proto)
	}
}

func (s *Server) startTCP() {
	var l net.Listener
	var err error

	switch s.Type {
	case "aes":
		l, err = StartAESListener(s.addr, s.secret)
	case "tls":
		l, err = StartTLSListener(s.addr, s.caFile, s.certFile, s.keyFile)
	default:
		l, err = StartDefaultListener(s.addr)
	}

	if err != nil {
		logrus.Errorf("start (%s) server on %s failed: %s", s.Type, s.addr, err)
		return
	}

	logrus.Infof("start (%s) server on %s success", s.Type, s.addr)

	for {
		rawConn, err := l.Accept()
		if err != nil {
			logrus.Errorf("accept new conn error: %s", err)
			continue // TODO: fix me!
		}

		logrus.Debugf("accept new client from %s", rawConn.RemoteAddr())
		var conn io.ReadWriteCloser
		hConn := emsg.NewConn(rawConn)

		if s.Type == "aes" {
			// TODO: custom cipher for each connection
			cipher := ecrypt.NewCipher("aes256cfb", []byte(s.secret))
			conn = connection.NewConn(rawConn, cipher)
			hConn.SetCipher(cipher)
		} else {
			conn = rawConn
		}

		// Important!
		rawConn.SetReadDeadline(time.Now().Add(time.Second * 6))

		if err := handshake(hConn); err != nil {
			logrus.Errorf("handshake failed: %s", err)
			conn.Close()
			continue
		}

		// Important! cancel timeout!
		rawConn.SetReadDeadline(time.Time{})

		go s.handleTCPClient(conn)
	}
}

func (s *Server) handleTCPClient(conn io.ReadWriteCloser) {
	// client_name := conn.((*net.TCPConn)).RemoteAddr()
	l := link.NewLink(&link.LinkConfig{IsServerSide: true})
	errCh := l.Join(conn)
	<-errCh

	// client 断开连接
	logrus.Warnf("client %#v is offline", conn)
	l.Close()
	conn.Close()
}

func (s *Server) startKCP() {
	config := &KCPServerConfig{
		Crypt:        "aes",
		Mode:         "fast",
		MTU:          1350,
		SndWnd:       1024,
		RcvWnd:       1024,
		DataShard:    10,
		ParityShard:  3,
		DSCP:         0,
		NoComp:       false,
		AckNodelay:   true,
		NoDelay:      0,
		Interval:     40,
		Resend:       0,
		NoCongestion: 0,
		SockBuf:      4194304,
		KeepAlive:    10,
	}

	// var l net.Listener
	// var err error

	pass := pbkdf2.Key([]byte(s.secret), []byte(KCP_SALT), 4096, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(pass)

	l, err := kcp.ListenWithOptions(s.addr, block, config.DataShard, config.ParityShard)
	if err != nil {
		logrus.Error("start kcp listen error:", err)
		return
	}
	logrus.Debugf("start kcp server listen on %s", l.Addr())

	if err := l.SetDSCP(config.DSCP); err != nil {
		logrus.Println("SetDSCP:", err)
	}
	if err := l.SetReadBuffer(config.SockBuf); err != nil {
		logrus.Println("SetReadBuffer:", err)
	}
	if err := l.SetWriteBuffer(config.SockBuf); err != nil {
		logrus.Println("SetWriteBuffer:", err)
	}

	for {
		rawConn, err := l.AcceptKCP()
		if err != nil {
			logrus.Errorf("accept new client error: %+v", err)
			continue // TODO: fix me!
		}

		logrus.Debugf("accept new client from %s", rawConn.RemoteAddr())

		rawConn.SetStreamMode(true)
		rawConn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
		rawConn.SetMtu(config.MTU)
		rawConn.SetWindowSize(config.SndWnd, config.RcvWnd)
		rawConn.SetACKNoDelay(config.AckNodelay)
		rawConn.SetKeepAlive(config.KeepAlive)

		go s.handleKCPClient(rawConn)
	}
}

func (s *Server) handleKCPClient(conn net.Conn) {
	client_name := conn.RemoteAddr()
	l := link.NewLink(nil)
	errCh := l.Join(conn)
	<-errCh

	// client 断开连接
	logrus.Warnf("client %s is offline", client_name)
	l.Close()
	conn.Close()
}
