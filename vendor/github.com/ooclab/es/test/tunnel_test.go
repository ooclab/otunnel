package test

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	setupChannelPing()
}

func setupChannelPing() error {
	_, clientLink, _ := getServerAndClient()
	go runPingServer("127.0.0.1:12345")
	localHost := "127.0.0.1"
	localPort := 12345
	remoteHost := "127.0.0.1"
	remotePort := 54321
	reverse := true
	if err := clientLink.OpenTunnel("tcp", localHost, localPort, remoteHost, remotePort, reverse); err != nil {
		logrus.Errorf("OpenTunnel failed: %s", err)
		return err
	}

	return nil
}

func Test_LinkChannelPing(t *testing.T) {
	if err := runPingClient("http://127.0.0.1:54321/ping"); err != nil {
		logrus.Errorf("runPingClient failed: %s", err)
		return
	}
}

func Test_LinkPing(t *testing.T) {
	_, clientLink, _ := getServerAndClient()
	_, err := clientLink.Ping()
	if err != nil {
		t.Errorf("ping error: %s", err)
	}
}

func Benchmark_LinkPing(b *testing.B) {
	_, clientLink, _ := getServerAndClient()
	for i := 0; i < b.N; i++ { //use b.N for looping
		clientLink.Ping()
	}
}

func Benchmark_LinkChannelPing(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		if err := runPingClient("http://127.0.0.1:54321/ping"); err != nil {
			logrus.Errorf("runPingClient failed: %s", err)
			return
		}
	}
}
