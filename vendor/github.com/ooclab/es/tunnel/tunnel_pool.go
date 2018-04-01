package tunnel

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
)

type Pool struct {
	curID     uint32
	idMutex   *sync.Mutex
	pool      map[uint32]*Tunnel
	poolMutex *sync.Mutex
}

func NewPool(isServerSide bool) *Pool {
	p := &Pool{
		idMutex:   &sync.Mutex{},
		pool:      map[uint32]*Tunnel{},
		poolMutex: &sync.Mutex{},
	}
	if isServerSide {
		p.curID = 1
	} else {
		p.curID = 2
	}
	return p
}

func (p *Pool) newID() uint32 {
	p.idMutex.Lock()
	defer p.idMutex.Unlock()
	for {
		p.curID += 2
		if p.curID <= 0 {
			continue
		}
		if !p.Exist(p.curID) {
			break
		}
	}
	return p.curID
}

func (p *Pool) Exist(id uint32) bool {
	p.poolMutex.Lock()
	_, exist := p.pool[id]
	p.poolMutex.Unlock()
	return exist
}

func (p *Pool) Get(id uint32) *Tunnel {
	p.poolMutex.Lock()
	v, exist := p.pool[id]
	p.poolMutex.Unlock()
	if exist {
		return v
	} else {
		return nil
	}
}

func (p *Pool) Delete(t *Tunnel) error {
	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()
	id := t.ID
	_, exist := p.pool[id]
	if !exist {
		return errors.New("delete failed: tunnel not exist")
	}
	delete(p.pool, id)
	return nil
}

func (p *Pool) New(manager *Manager, cfg *TunnelConfig) (*Tunnel, error) {
	if cfg.ID == 0 {
		cfg.ID = p.newID()
	} else {
		if p.Exist(cfg.ID) {
			logrus.Errorf("create tunnel by id failed: id %d is existed", cfg.ID)
			return nil, errors.New("tunnel ID existed")
		}
	}

	p.poolMutex.Lock()
	t := newTunnel(manager, cfg)
	p.pool[cfg.ID] = t
	p.poolMutex.Unlock()

	return t, nil
}
