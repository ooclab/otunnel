package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ooclab/es"
	"github.com/ooclab/es/ecrypt"
	"github.com/ooclab/es/link"
	"github.com/ooclab/otunnel/pkg/util"
)

// StartDefaultConnect start connection to a default server
func StartDefaultConnect(addr string) (net.Conn, error) {
	rawConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return rawConn, nil
}

// StartAESConnect start connection to a aes server
func StartAESConnect(addr string, secret []byte) (net.Conn, error) {
	return StartDefaultConnect(addr)
}

// StartTLSConnect start connection to a tls server
func StartTLSConnect(addr string, caFile string, certFile string, keyFile string) (net.Conn, error) {
	conn, err := tls.Dial(
		"tcp",
		addr,
		&tls.Config{
			InsecureSkipVerify: false,
			ServerName:         "otunnelDefaultServer",
		},
	)
	if err != nil {
		logrus.Errorf("connect to %s failed: %s", addr, err)
		return nil, err
	}

	return conn, nil
}

// Client define a client object
type Client struct {
	Proto   string
	Type    string
	link    *link.Link
	tunnels []string

	addr string

	// aes connection needed!
	secret []byte

	keepaliveInterval time.Duration

	// tls connection needed!
	caFile   string
	keyFile  string
	certFile string
}

// NewClient create a server object
func newClient(c *cli.Context) (*Client, error) {
	if c.NArg() == 0 {
		return nil, errors.New("NEED server address")
	}

	// TODO: more server support!
	addr := c.Args().First()

	client := &Client{
		Proto:             c.String("proto"),
		addr:              addr,
		secret:            util.GenSecret(c.String("secret"), c.Int("keyiter"), c.Int("keylen")),
		keepaliveInterval: time.Duration(c.Int("keepalive")) * time.Second,
		caFile:            c.String("ca"),
		certFile:          c.String("cert"),
		keyFile:           c.String("key"),
		tunnels:           c.StringSlice("tunnel"),
	}

	if len(client.secret) > 0 {
		client.Type = "aes"
	} else if len(client.certFile) > 0 && len(client.keyFile) > 0 {
		client.Type = "tls"
	} else {
		client.Type = "default"
	}

	return client, nil
}

func (client *Client) connect() (es.Conn, error) {
	switch client.Proto {
	case "tcp":
		return client.connectTCP()
	default:
		logrus.Errorf("unknown proto : %s", client.Proto)
		return nil, errors.New("unknown link proto")
	}
}

func (client *Client) connectTCP() (es.Conn, error) {
	// logrus.Debugf("connect to %s", client.addr)

	var rawConn net.Conn
	var err error

	switch client.Type {
	case "tls":
		rawConn, err = StartTLSConnect(client.addr, client.caFile, client.certFile, client.keyFile)
	case "aes":
		rawConn, err = StartAESConnect(client.addr, client.secret)
	default:
		rawConn, err = StartDefaultConnect(client.addr)
	}

	if err != nil {
		logrus.Errorf("connect to %s failed: %s", client.addr, err)
		return nil, err
	}

	logrus.Debugf("connect to %s success", rawConn.RemoteAddr())

	var conn es.Conn

	if client.Type == "aes" {
		// TODO: custom cipher for each connection
		cipher := ecrypt.NewCipher("aes256cfb", client.secret)
		conn = es.NewSafeConn(rawConn, cipher)
	} else {
		conn = es.NewBaseConn(rawConn)
	}

	if err := handshake(conn); err != nil {
		logrus.Errorf("handshake failed: %s", err)
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// Start run a server
func (client *Client) Start() {
	switch client.Proto {
	case "tcp":
		client.startTCP()
	default:
		logrus.Errorf("unknown proto : %s", client.Proto)
	}
}

func (client *Client) startTCP() {

	for {
		conn, err := client.connect()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		l := link.NewLink(&link.LinkConfig{
			IsServerSide:      false,
			KeepaliveInterval: client.keepaliveInterval,
		})
		l.Bind(conn)
		defer l.Close()
		for _, t := range client.tunnels {
			proto, localHost, localPort, remoteHost, remotePort, reverse, err := parseTunnel(t)
			if err != nil {
				logrus.Fatalf("parse tunnel failed: %s", err)
			}
			l.OpenTunnel(proto, localHost, localPort, remoteHost, remotePort, reverse)
		}

		l.Wait()
		time.Sleep(1 * time.Second) // TODO: sleep smartly
	}

}

func parseTunnel(value string) (proto string, localHost string, localPort int, remoteHost string, remotePort int, reverse bool, err error) {
	L := strings.Split(value, ":")

	// !IMPORTANT! support old configure
	if len(L) == 5 {
		L = append([]string{L[0], "tcp"}, L[1:]...)
	}

	if len(L) != 6 {
		fmt.Println("tunnel format: \"r|f:proto:local_host:local_port:remote_host:remote_port\"")
		err = errors.New("tunnel map is wrong: " + value)
		return
	}

	switch L[0] {
	case "r", "R":
		reverse = true
	case "f", "F":
		reverse = false
	default:
		err = errors.New("wrong tunnel map")
		return
	}

	proto = strings.TrimSpace(strings.ToLower(L[1]))
	if proto == "" {
		proto = "tcp"
	}
	if !(proto == "tcp" || proto == "udp") {
		err = errors.New("unknown protocol")
		return
	}
	localHost = L[2]
	remoteHost = L[4]
	localPort, err = strconv.Atoi(L[3])
	if err != nil {
		return
	}
	remotePort, err = strconv.Atoi(L[5])
	if err != nil {
		return
	}

	return
}
