package link

import (
	"encoding/json"
	"errors"

	"github.com/ooclab/es/session"
	"github.com/ooclab/es/tunnel"
	"github.com/sirupsen/logrus"
)

func defaultOpenTunnel(sessionManager *session.Manager, tunnelManager *tunnel.Manager) OpenTunnelFunc {
	return func(proto string, localHost string, localPort int, remoteHost string, remotePort int, reverse bool) error {
		// send open tunnel message to remote endpoint
		cfg := &tunnel.TunnelConfig{
			LocalHost:  localHost,
			LocalPort:  localPort,
			RemoteHost: remoteHost,
			RemotePort: remotePort,
			Reverse:    reverse,
			Proto:      proto,
		}

		body, _ := json.Marshal(cfg.RemoteConfig())
		s, err := sessionManager.New()
		if err != nil {
			logrus.WithField("error", err).Error("open session failed")
			return err
		}

		resp, err := s.SendAndWait(&session.Request{
			Action: "/tunnel",
			Body:   body,
		})
		if err != nil {
			logrus.WithField("error", err).Error("send request to remote endpoint failed")
			return err
		}

		// fmt.Println("resp: ", resp)
		if resp.Status != "success" {
			logrus.WithField("error", err).Error("open tunnel in the remote endpoint failed")
			return errors.New("open tunnel in the remote endpoint failed")
		}

		tcBody := tunnelCreateBody{}
		if err = json.Unmarshal(resp.Body, &tcBody); err != nil {
			logrus.WithField("error", err).Error("json unmarshal body failed")
			return errors.New("json unmarshal body error")
		}

		// success: open tunnel at local endpoint
		logrus.WithField("config", cfg).Debug("open tunnel in the remote endpoint success")

		cfg.ID = tcBody.ID
		t, err := tunnelManager.TunnelCreate(cfg)
		if err != nil {
			logrus.Errorf("open tunnel in the local side failed: %s", err)
			// TODO: close the tunnel in remote endpoint
			return errors.New("open tunnel in the local side failed")
		}

		logrus.WithField("tunnel", t).Debug("open tunnel in the local side success")

		return nil
	}
}
