package client

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	_ "net/http/pprof"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xtaci/kcp-go"

	"github.com/ooclab/es/emsg"
	"github.com/ooclab/es/link"
)

// Config for client
type KCPClientConfig struct {
	Key   string `json:"key"`
	Crypt string `json:"crypt"`
	Mode  string `json:"mode"`

	Conn       int `json:"conn"`
	AutoExpire int `json:"autoexpire"`

	MTU          int  `json:"mtu"`
	SndWnd       int  `json:"sndwnd"`
	RcvWnd       int  `json:"rcvwnd"`
	DataShard    int  `json:"datashard"`
	ParityShard  int  `json:"parityshard"`
	DSCP         int  `json:"dscp"`
	NoComp       bool `json:"nocomp"`
	AckNodelay   bool `json:"acknodelay"`
	NoDelay      int  `json:"nodelay"`
	Interval     int  `json:"interval"`
	Resend       int  `json:"resend"`
	NoCongestion int  `json:"nc"`
	SockBuf      int  `json:"sockbuf"`
	KeepAlive    int  `json:"keepalive"`
}

var (
	// SALT is use for pbkdf2 key expansion
	KCP_SALT = "kcp-go"
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

	pprof := c.String("pprof")
	if pprof != "" {
		go func() {
			logrus.Println(http.ListenAndServe(pprof, nil))
		}()
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

func (client *Client) connect() (net.Conn, error) {
	switch client.Proto {
	case "tcp":
		return client.connectTCP()
	// case "kcp":
	// 	return client.connectKCP()
	default:
		logrus.Errorf("unknown proto : %s", client.Proto)
		return nil, errors.New("unknown link proto")
	}
}

func (client *Client) connectTCP() (net.Conn, error) {
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

	conn := emsg.NewConn(rawConn)
	if client.Type == "aes" {
		logrus.Warn("aes is not completed")
		// TODO: custom cipher for each connection
		// cipher := emsg.NewCipher("aes256cfb", []byte(client.secret))
		// conn.SetCipher(cipher)
	}

	if err := handshake(conn); err != nil {
		logrus.Errorf("handshake failed: %s", err)
		conn.Close()
		return nil, err
	}

	return rawConn, nil
}

func (client *Client) connectKCP() (*kcp.UDPSession, error) {
	logrus.Debugf("connect to KCP server %s", client.addr)

	config := &KCPClientConfig{
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

	pass := pbkdf2.Key([]byte(client.secret), []byte(KCP_SALT), 4096, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(pass)

	kcpconn, err := kcp.DialWithOptions(client.addr, block, config.DataShard, config.ParityShard)
	if err != nil {
		logrus.Errorf("connect to kcp server %s error: %s", client.addr, err)
		return nil, err
	}

	kcpconn.SetStreamMode(true)
	kcpconn.SetNoDelay(config.NoDelay, config.Interval, config.Resend, config.NoCongestion)
	kcpconn.SetWindowSize(config.SndWnd, config.RcvWnd)
	kcpconn.SetMtu(config.MTU)
	kcpconn.SetACKNoDelay(config.AckNodelay)
	kcpconn.SetKeepAlive(config.KeepAlive)

	if err := kcpconn.SetDSCP(config.DSCP); err != nil {
		logrus.Error("SetDSCP:", err)
	}
	if err := kcpconn.SetReadBuffer(config.SockBuf); err != nil {
		logrus.Error("SetReadBuffer:", err)
	}
	if err := kcpconn.SetWriteBuffer(config.SockBuf); err != nil {
		logrus.Error("SetWriteBuffer:", err)
	}

	logrus.Debugf("connect to kcp server %s success", kcpconn.RemoteAddr())

	// conn := emsg.NewConn(kcpconn)
	return kcpconn, nil
}

// Start run a server
func (client *Client) Start() {
	switch client.Proto {
	case "tcp":
		client.startTCP()
	case "kcp":
		client.startKCP()
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

		l := link.NewLink(nil)
		l.Bind(conn)
		for _, t := range client.tunnels {
			localHost, localPort, remoteHost, remotePort, reverse, err := parseTunnel(t)
			if err != nil {
				panic(err)
			}
			l.OpenTunnel(localHost, localPort, remoteHost, remotePort, reverse)
		}
		l.WaitDisconnected()
		l.Close()
		time.Sleep(1 * time.Second) // TODO: sleep smartly
	}

}

func (client *Client) startKCP() {

	for {
		l := link.NewLink(nil)

		conn, err := client.connect()
		if err != nil {
			logrus.Errorf("start connect failed: %s", err)
			return
		}

		l.Bind(conn)
		// l.OpenTunnel(localHost, localPort, remoteHost, remotePort, reverse)
		l.WaitDisconnected()

		//  FIXME! 不应该到这里！
		logrus.Debugf("client connection break!")
		time.Sleep(1 * time.Second)
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
