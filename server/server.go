package server

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es"
	"github.com/ooclab/es/ecrypt"
	"github.com/ooclab/es/link"
	"github.com/urfave/cli"
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

	keepaliveInterval time.Duration

	// tls connection needed!
	caFile   string
	keyFile  string
	certFile string
}

func newServer(c *cli.Context) *Server {
	addr := c.Args().First()
	if len(addr) == 0 {
		addr = ":10000"
	}

	s := &Server{
		Proto:             c.String("proto"),
		addr:              addr,
		secret:            c.String("secret"),
		keepaliveInterval: time.Duration(c.Int("keepalive")) * time.Second,
		caFile:            c.String("ca"),
		certFile:          c.String("cert"),
		keyFile:           c.String("key"),
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
		var conn es.Conn

		if s.Type == "aes" {
			// TODO: custom cipher for each connection
			cipher := ecrypt.NewCipher("aes256cfb", []byte(s.secret))
			conn = es.NewSafeConn(rawConn, cipher)
		} else {
			conn = es.NewBaseConn(rawConn)
		}

		// Important!
		rawConn.SetReadDeadline(time.Now().Add(time.Second * 6))

		if err := handshake(conn); err != nil {
			logrus.Errorf("handshake failed: %s", err)
			conn.Close()
			continue
		}

		// Important! cancel timeout!
		rawConn.SetReadDeadline(time.Time{})

		go s.handleTCPClient(conn)
	}
}

func (s *Server) handleTCPClient(conn es.Conn) {
	// client_name := conn.((*net.TCPConn)).RemoteAddr()
	l := link.NewLink(&link.LinkConfig{
		IsServerSide:      true,
		KeepaliveInterval: s.keepaliveInterval,
	})
	l.Bind(conn)
	l.Wait()
	logrus.Warnf("client %#v is offline", conn)
	l.Close()
	conn.Close()
}
