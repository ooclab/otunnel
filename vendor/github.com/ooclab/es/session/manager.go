package session

import (
	"errors"

	"github.com/ooclab/es"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	pool           *Pool
	outbound       chan []byte
	requestHandler RequestHandler
}

func NewManager(isServerSide bool, outbound chan []byte) *Manager {
	m := &Manager{
		pool:     newPool(isServerSide),
		outbound: outbound,
	}
	return m
}

func (manager *Manager) SetRequestHandler(hdr RequestHandler) {
	manager.requestHandler = hdr
}

func (manager *Manager) HandleIn(payload []byte) error {
	m, err := LoadEMSG(payload)
	if err != nil {
		return err
	}

	switch m.Type {

	case MsgTypeRequest:
		rMsg := manager.requestHandler.Handle(m)
		manager.outbound <- append([]byte{es.LinkMsgTypeSession}, rMsg.Bytes()...)

	case MsgTypeResponse:
		s := manager.pool.Get(m.ID)
		if s == nil {
			logrus.Errorf("can not find session with ID %d", m.ID)
			return errors.New("no such session")
		}
		s.HandleResponse(m.Payload)

	default:
		logrus.Errorf("unknown session msg type: %d", m.Type)
		return errors.New("unknown session msg type")

	}

	return nil
}

func (manager *Manager) New() (*Session, error) {
	return manager.pool.New(manager.outbound)
}

func (manager *Manager) Close() {
	for item := range manager.pool.IterBuffered() {
		item.Val.Close()
		logrus.WithFields(logrus.Fields{
			"session_id": item.Key,
		}).Debug("close session")
		manager.pool.Delete(item.Val)
	}
}
