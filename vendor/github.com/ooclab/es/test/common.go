package test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/ooclab/es"
	"github.com/ooclab/es/link"
	"github.com/sirupsen/logrus"
)

// For test
// func init() {
// 	logrus.SetFormatter(&logrus.TextFormatter{
// 		FullTimestamp:   true,
// 		TimestampFormat: "01/02 15:04:05",
// 	})
// 	logrus.SetLevel(logrus.DebugLevel)
// }

func getServerAndClient() (serverLink *link.Link, clientLink *link.Link, err error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logrus.Error("listen error:", err)
		return
	}

	port := l.Addr().(*net.TCPAddr).Port
	// fmt.Printf("start listen on %d\n", port)

	serverLinkCh := make(chan *link.Link, 1)

	// run server
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				logrus.Error("accept new conn error: ", err)
				continue // TODO: fix me!
			}

			go func() {
				l := link.NewLink(nil)
				serverLinkCh <- l
				ec := es.NewBaseConn(conn)
				l.Bind(ec)
				l.Close()
				fmt.Println("link quit: ", err)
			}()
		}
	}()

	// run client
	clientLink = connectServer(fmt.Sprintf("127.0.0.1:%d", port))
	serverLink = <-serverLinkCh

	return serverLink, clientLink, nil
}

func tcpConnect(addrS string) (conn net.Conn, err error) {
	// fmt.Printf("try connect to relay server: %s\n", addr_s)
	addr, err := net.ResolveTCPAddr("tcp", addrS)
	if err != nil {
		fmt.Printf("resolve relay-server (%s) failed: %s", addr, err)
		return
	}
	conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		logrus.Errorf("dial %s failed: %s", addr, err.Error())
		return
	}
	// fmt.Printf("connect to relay server %s success\n", conn.RemoteAddr())

	return
}

func connectServer(addr string) *link.Link {
	conn, err := tcpConnect(addr)
	if err != nil {
		panic(err)
	}
	l := link.NewLink(nil)
	ec := es.NewBaseConn(conn)
	// FIXME: quit it not a good choice for testcase!
	go func() {
		l.Bind(ec)
		fmt.Println("link quit: ", err)
		l.Close()
	}()
	return l
}

func runPingServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	http.ListenAndServe(addr, mux)
}

func runPingClient(addr string) error {
	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logrus.Debug("respBody = ", string(respBody))
	if string(respBody) != "pong" {
		return errors.New("not-pong")
	}
	return nil
}
