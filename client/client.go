package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ooclab/es/ecrypt"
	"github.com/ooclab/es/link"
	"github.com/ooclab/otunnel/common/connection"
	"github.com/ooclab/otunnel/common/emsg"
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
func StartAESConnect(addr string, secret string) (net.Conn, error) {
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
	secret string

	// tls connection needed!
	caFile   string
	keyFile  string
	certFile string
}

// NewClient create a server object
func NewClient(c *cli.Context) (*Client, error) {
	if c.NArg() == 0 {
		return nil, errors.New("NEED server address")
	}

	// TODO: more server support!
	addr := c.Args().First()

	client := &Client{
		Proto:    c.String("proto"),
		addr:     addr,
		secret:   c.String("secret"),
		caFile:   c.String("ca"),
		certFile: c.String("cert"),
		keyFile:  c.String("key"),
		tunnels:  c.StringSlice("tunnel"),
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

func (client *Client) connect() (io.ReadWriteCloser, error) {
	switch client.Proto {
	case "tcp":
		return client.connectTCP()
	default:
		logrus.Errorf("unknown proto : %s", client.Proto)
		return nil, errors.New("unknown link proto")
	}
}

func (client *Client) connectTCP() (io.ReadWriteCloser, error) {
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

	var conn io.ReadWriteCloser
	hConn := emsg.NewConn(rawConn)

	if client.Type == "aes" {
		// TODO: custom cipher for each connection
		cipher := ecrypt.NewCipher("aes256cfb", []byte(client.secret))
		conn = connection.NewConn(rawConn, cipher)
		hConn.SetCipher(cipher)
	} else {
		conn = rawConn
	}

	if err := handshake(hConn); err != nil {
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

		l := link.NewLink(&link.LinkConfig{IsServerSide: false})
		errCh := l.Join(conn)
		for _, t := range client.tunnels {
			localHost, localPort, remoteHost, remotePort, reverse, err := parseTunnel(t)
			if err != nil {
				panic(err)
			}
			l.OpenTunnel(localHost, localPort, remoteHost, remotePort, reverse)
		}
		<-errCh
		l.Close()
		time.Sleep(1 * time.Second) // TODO: sleep smartly
	}

}

func parseTunnel(value string) (localHost string, localPort int, remoteHost string, remotePort int, reverse bool, err error) {
	L := strings.Split(value, ":")
	if len(L) != 5 {
		fmt.Println("tunnel format: \"r|f:local_host:local_port:remote_host:remote_port\"")
		err = errors.New("tunnel map is wrong: " + value)
		return
	}

	localHost = L[1]
	remoteHost = L[3]
	switch L[0] {
	case "r", "R":
		reverse = true
	case "f", "F":
		reverse = false
	default:
		err = errors.New("wrong tunnel map")
		return
	}
	localPort, err = strconv.Atoi(L[2])
	if err != nil {
		return
	}
	remotePort, err = strconv.Atoi(L[4])
	if err != nil {
		return
	}

	return
}
