package es

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es/ecrypt"
)

func testEcho(conn Conn) error {
	maxSize := 1024 * 64
	var start time.Time
	var td time.Duration
	var speed float64
	for i := 2; i <= maxSize; i = i * 2 {
		b := make([]byte, i-1)
		rand.Read(b)
		sc := md5.Sum(b)

		// Send
		start = time.Now()
		if err := conn.Send(b); err != nil {
			return err
		}
		td = time.Since(start)
		speed = (float64(i+1) / td.Seconds()) / (1024 * 1024)
		msg, err := conn.Recv()
		if err != nil {
			return err
		}

		rc := md5.Sum(msg)
		if rc != sc {
			logrus.Errorf("%d msg is mismatch", i+1)
			return errors.New("echo mismatch")
		}
		logrus.Infof("%12d %16s %16f M/s", i+1, td, speed)
	}
	return conn.Send([]byte("quit"))
}

func Test_BaseConn(t *testing.T) {
	// run server
	l, _ := net.Listen("tcp", "127.0.0.1:")
	go func() {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		bc := NewBaseConn(conn)
		defer bc.Close()
		for {
			msg, err := bc.Recv()
			if err != nil {
				t.Errorf("server Recv failed: %s", err)
				break
			}
			if len(msg) == 4 && string(msg) == "quit" {
				break
			}
			bc.Send(msg)
		}
	}()

	// run client
	conn, _ := net.Dial("tcp", l.Addr().String())
	bc := NewBaseConn(conn)
	defer bc.Close()

	testEcho(bc)
}

func Test_SafeConn(t *testing.T) {
	secret := []byte("longlongsecret")
	// run server
	l, _ := net.Listen("tcp", "127.0.0.1:")
	go func() {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		cipher := ecrypt.NewCipher("aes256cfb", secret)
		bc := NewSafeConn(conn, cipher)
		defer bc.Close()
		for {
			msg, err := bc.Recv()
			if err != nil {
				t.Errorf("server Recv failed: %s", err)
				break
			}
			if len(msg) == 4 && string(msg) == "quit" {
				break
			}
			bc.Send(msg)
		}
	}()

	// run client
	conn, _ := net.Dial("tcp", l.Addr().String())
	cipher := ecrypt.NewCipher("aes256cfb", secret)
	sc := NewSafeConn(conn, cipher)
	defer sc.Close()

	testEcho(sc)
}
