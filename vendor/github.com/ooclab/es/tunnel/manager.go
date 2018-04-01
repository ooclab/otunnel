package tunnel

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/ooclab/es/session"
	tcommon "github.com/ooclab/es/tunnel/common"
)

var globalListenPool = newListenPool()

type Manager struct {
	pool           *Pool
	lpool          *listenPool
	outbound       chan []byte
	sessionManager *session.Manager
}

func NewManager(isServerSide bool, outbound chan []byte, sm *session.Manager) *Manager {
	return &Manager{
		pool:           NewPool(isServerSide),
		lpool:          globalListenPool,
		outbound:       outbound,
		sessionManager: sm,
	}
}

func (manager *Manager) HandleIn(payload []byte) error {
	m, err := tcommon.LoadTMSG(payload)
	if err != nil {
		return err
	}

	switch m.Type {

	case tcommon.MsgTypeChannelForward:
		t := manager.pool.Get(m.TunnelID)
		if t == nil {
			logrus.Warnf("can not find tunnel %d", m.TunnelID)
			return errors.New("can not find tunnel")
		}
		return t.HandleIn(m)

	case tcommon.MsgTypeChannelClose:
		t := manager.pool.Get(m.TunnelID)
		if t == nil {
			logrus.Warnf("can not find tunnel %d", m.TunnelID)
			return errors.New("no such tunnel")
		}
		t.HandleChannelClose(m)
		// return nil

	default:
		logrus.Errorf("unknown tunnel msg type: %d", m.Type)
		return errors.New("unknown tunnel msg type")

	}

	return nil
}

func (manager *Manager) TunnelCreate(cfg *TunnelConfig) (*Tunnel, error) {
	logrus.Debugf("prepare to create a tunnel with config %+v", cfg)
	t, err := manager.pool.New(manager, cfg)
	if err != nil {
		logrus.Errorf("create new tunnel failed: %s", err)
		return nil, err
	}

	if err := t.Listen(); err != nil {
		logrus.Errorf("run forward tunnel %s failed: %s", t.String(), err)
		manager.pool.Delete(t)
		return nil, err
	}

	logrus.Debugf("create forward tunnel: %+v", t)
	return t, nil
}

func (manager *Manager) Close() error {
	for item := range manager.lpool.IterBuffered() {
		manager.lpool.Delete(item.Key)
	}
	return nil
}
