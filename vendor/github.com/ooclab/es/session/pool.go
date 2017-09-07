package session

import (
	"errors"
	"sync"
)

type Pool struct {
	curID     uint32
	idMutex   *sync.Mutex
	pool      map[uint32]*Session
	poolMutex *sync.Mutex
}

func newPool(isServerSide bool) *Pool {
	p := &Pool{
		idMutex:   &sync.Mutex{},
		pool:      map[uint32]*Session{},
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

func (p *Pool) New(outbound chan []byte) (*Session, error) {
	id := p.newID()

	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()

	session := newSession(id, outbound)
	p.pool[id] = session

	return session, nil
}

func (p *Pool) Get(id uint32) *Session {
	p.poolMutex.Lock()
	v, exist := p.pool[id]
	p.poolMutex.Unlock()
	if exist {
		return v
	} else {
		return nil
	}
}

func (p *Pool) Delete(session *Session) error {
	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()

	_, exist := p.pool[session.ID]
	if !exist {
		return errors.New("delete failed: session not exist")
	}
	delete(p.pool, session.ID)
	return nil
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type PoolTuple struct {
	Key uint32
	Val *Session
}

// Returns a buffered iterator which could be used in a for range loop.
func (p *Pool) IterBuffered() <-chan PoolTuple {
	ch := make(chan PoolTuple, len(p.pool))
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			// Foreach key, value pair.
			p.poolMutex.Lock()
			defer p.poolMutex.Unlock()
			for key, val := range p.pool {
				ch <- PoolTuple{key, val}
			}
			wg.Done()
		}()
		wg.Wait()
		close(ch)
	}()
	return ch
}
