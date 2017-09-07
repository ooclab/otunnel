package channel

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

type Pool struct {
	nextID    uint32
	pool      map[uint32]Channel
	poolMutex sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{
		nextID:    1,
		pool:      map[uint32]Channel{},
		poolMutex: sync.RWMutex{},
	}
}

func (p *Pool) newID() (id uint32) {
	for {
		id = atomic.AddUint32(&p.nextID, 1)
		if !p.Exist(id) {
			return
		}
	}
}

func (p *Pool) Exist(id uint32) bool {
	p.poolMutex.RLock()
	_, exist := p.pool[id]
	p.poolMutex.RUnlock()
	return exist
}

func (p *Pool) Get(id uint32) Channel {
	p.poolMutex.Lock()
	v, exist := p.pool[id]
	p.poolMutex.Unlock()
	if exist {
		return v
	} else {
		return nil
	}
}

func (p *Pool) Delete(c Channel) error {
	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()

	_, exist := p.pool[c.ID()]
	if !exist {
		return errors.New("delete failed: channel not exist")
	}
	c.Close() // FIXME!
	delete(p.pool, c.ID())
	return nil
}

func (p *Pool) New(tid uint32, outbound chan []byte, conn net.Conn) Channel {
	cid := p.newID()
	return p.NewByID(cid, tid, outbound, conn)
}

func (p *Pool) NewByID(cid uint32, tid uint32, outbound chan []byte, conn net.Conn) Channel {
	p.poolMutex.Lock()
	c := &tcpChannel{
		tid:      tid,
		cid:      cid,
		outbound: outbound,
		conn:     conn,
		lock:     &sync.Mutex{},
	}
	p.pool[cid] = c
	p.poolMutex.Unlock()
	return c
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type poolTuple struct {
	Key uint32
	Val Channel
}

// Returns a buffered iterator which could be used in a for range loop.
func (p *Pool) IterBuffered() <-chan poolTuple {
	ch := make(chan poolTuple, len(p.pool))
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			// Foreach key, value pair.
			p.poolMutex.Lock()
			defer p.poolMutex.Unlock()
			for key, val := range p.pool {
				ch <- poolTuple{key, val}
			}
			wg.Done()
		}()
		wg.Wait()
		close(ch)
	}()
	return ch
}
