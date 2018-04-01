package udp

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	// logrus.SetLevel(logrus.DebugLevel)
}

func runServer(quit chan struct{}) (net.Addr, error) {
	addr := &net.UDPAddr{
		// Port: 1234,
		IP: net.ParseIP("127.0.0.1"),
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("net.ListendUDP error: %v\n", err)
		return nil, err
	}
	sock, err := NewServerSocket(conn)
	if err != nil {
		fmt.Println("udp.NewServerSocket error: ", err)
		return nil, err
	}

	// quit
	go func() {
		<-quit
		conn.Close()
	}()

	go func() {
		for {
			conn, err := sock.Accept()
			if err != nil {
				fmt.Println("accept failed: ", err)
				break
			}
			// fmt.Println("accept client: ", conn)
			go func() {
				// fmt.Println("start client recv: ", conn)
				// defer func() { fmt.Println("quit conn: ", conn) }()
				for {
					msg, err := conn.RecvMsg()
					if err != nil {
						fmt.Printf("recv msg failed: %s\n", err)
						return
					}
					rc := md5.Sum(msg)
					logrus.Debugf("%9d <-- recv %s\n", len(msg), hex.EncodeToString(rc[:]))
					err = conn.SendMsg(msg)
					if err != nil {
						fmt.Println("send msg failed: ", err)
						return
					}
				}
			}()
		}
	}()

	return conn.LocalAddr(), nil
}

func Test_Socket_RecvMsg_SendMsg(t *testing.T) {
	quit := make(chan struct{})
	defer close(quit)

	raddr, err := runServer(quit)
	if err != nil {
		t.Errorf("runServer failed: %s", err)
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}

	sock, clientConn, err := NewClientSocket(conn, raddr.(*net.UDPAddr))
	if err != nil {
		fmt.Printf("create client socket failed: %s", err)
		return
	}

	maxSize := 1024 * 1024 * 16
	var start time.Time
	var td time.Duration
	var speed float64
	for j := 0; j < 1; j++ {
		for i := 2; i <= maxSize; i = i * 2 {
			b := make([]byte, i+1)
			rand.Read(b)
			sc := md5.Sum(b)

			// Send
			start = time.Now()
			if err := clientConn.SendMsg(b); err != nil {
				logrus.Errorf("SendMsg failed: %s", err)
				break
			}
			td = time.Since(start)
			speed = (float64(i+1) / td.Seconds()) / (1024 * 1024)
			logrus.Debugf("%9d --> send %s %16s %16f M/s", i+1, hex.EncodeToString(sc[:]), td, speed)

			msg, err := clientConn.RecvMsg()
			if err != nil {
				t.Errorf("RecvMsg failed: %s", err)
				break
			}

			// Recv
			rc := md5.Sum(msg)
			logrus.Debugf("%9d <-- recv %s", len(msg), hex.EncodeToString(rc[:]))
			if md5.Sum(msg) != sc {
				t.Errorf("%d msg is mismatch", i+1)
				break
			}
		}
	}

	clientConn.Close()
	sock.Close()
}

func Test_Socket_WriterReaderCloser(t *testing.T) {
	quit := make(chan struct{})
	defer close(quit)

	raddr, err := runServer(quit)
	if err != nil {
		t.Errorf("runServer failed: %s", err)
	}

	maxSize := 1024*1024*128 + 369
	var start time.Time
	var td time.Duration
	var speed float64

	for j := 0; j < 1; j++ {
		func() {
			b := make([]byte, maxSize)
			// stop := make(chan struct{})
			// go func() {
			// 	fmt.Print("generate random buffer ")
			// 	for {
			// 		select {
			// 		case <-time.After(1 * time.Second):
			// 			fmt.Print(".")
			// 		case <-stop:
			// 			fmt.Println("")
			// 			return
			// 		}
			// 	}
			// }()
			rand.Read(b)
			// close(stop)
			sc := md5.Sum(b)

			// open client conn
			conn, err := net.ListenUDP("udp", nil)
			if err != nil {
				fmt.Printf("Some error %v", err)
				return
			}
			defer conn.Close()

			sock, clientConn, err := NewClientSocket(conn, raddr.(*net.UDPAddr))
			if err != nil {
				fmt.Printf("create client socket failed: %s", err)
				return
			}

			// send
			go func() {
				// Send
				start = time.Now()
				if _, err := clientConn.Write(b); err != nil {
					logrus.Errorf("SendMsg failed: %s", err)
					return
				}
				td = time.Since(start)
				speed = (float64(len(b)) / td.Seconds()) / (1024 * 1024)
				logrus.Debugf("%9d --> send %s %16s %16f M/s", len(b), hex.EncodeToString(sc[:]), td, speed)
			}()

			// recv
			rmsg := make([]byte, len(b))
			read := 0
			for read < len(b) {
				n, err := clientConn.Read(rmsg[read:])
				if err != nil {
					t.Errorf("RecvMsg failed: %s", err)
					break
				}
				// Recv
				rc := md5.Sum(rmsg[read : read+n])
				logrus.Debugf("Read: <-- recv %s", hex.EncodeToString(rc[:]))
				read += n
			}

			if md5.Sum(rmsg) != sc {
				t.Errorf("%d msg is mismatch", len(b))
				return
			}

			clientConn.Close()
			sock.Close()
		}()
	}
}
