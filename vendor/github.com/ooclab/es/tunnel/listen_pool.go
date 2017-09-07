package tunnel

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/Sirupsen/logrus"
)

type listenTarget struct {
	tunnel *Tunnel
	proto  string
	addr   string
	t      interface{}
	closed bool
	m      *sync.Mutex
}

func newTCPListenTarget(tunnel *Tunnel, host string, port int, l net.Listener) *listenTarget {
	return &listenTarget{
		tunnel: tunnel,
		proto:  "tcp",
		addr:   fmt.Sprintf("%s:%d", host, port),
		t:      l,
		m:      &sync.Mutex{},
	}
}

func newUDPListenTarget(tunnel *Tunnel, host string, port int, conn *net.UDPConn) *listenTarget {
	return &listenTarget{
		tunnel: tunnel,
		proto:  "udp",
		addr:   fmt.Sprintf("%s:%d", host, port),
		t:      conn,
	}
}

func (l *listenTarget) Close() error {
	// TODO: not use lock
	l.m.Lock()
	defer l.m.Unlock()

	if l.closed {
		return nil
	}

	defer func() {
		l.closed = true
	}()

	switch l.proto {
	case "tcp":
		logrus.Debugf("closing tcp listen: %s", l.addr)
		return l.t.(net.Listener).Close()
	case "udp":
		logrus.Debugf("closing udp listen: %s", l.addr)
		return l.t.(*net.UDPConn).Close()
	}
	return nil // FIXME!
}

type listenPool struct {
	pool      map[string]*listenTarget
	poolMutex *sync.Mutex
}

func newListenPool() *listenPool {
	return &listenPool{
		pool:      map[string]*listenTarget{},
		poolMutex: &sync.Mutex{},
	}
}

func (p *listenPool) TCPKey(host string, port int) string {
	return fmt.Sprintf("tcp:%s:%d", host, port)
}

func (p *listenPool) UDPKey(host string, port int) string {
	return fmt.Sprintf("tcp:%s:%d", host, port)
}

func (p *listenPool) Exist(key string) bool {
	p.poolMutex.Lock()
	_, exist := p.pool[key]
	p.poolMutex.Unlock()
	return exist
}

func (p *listenPool) Add(key string, value *listenTarget) {
	p.poolMutex.Lock()
	p.pool[key] = value
	p.poolMutex.Unlock()
}

func (p *listenPool) Get(key string) *listenTarget {
	p.poolMutex.Lock()
	v, exist := p.pool[key]
	p.poolMutex.Unlock()
	if exist {
		return v
	}

	return nil
}

func (p *listenPool) Delete(key string) error {
	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()

	v, exist := p.pool[key]
	if !exist {
		return errors.New("delete failed: listen port not exist")
	}
	v.Close() // FIXME!
	delete(p.pool, key)
	logrus.Debugf("delete listen %s from pool success", key)
	return nil
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type listenPoolTuple struct {
	Key string
	Val *listenTarget
}

// Returns a buffered iterator which could be used in a for range loop.
func (p *listenPool) IterBuffered() <-chan listenPoolTuple {
	ch := make(chan listenPoolTuple, len(p.pool))
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			// Foreach key, value pair.
			p.poolMutex.Lock()
			defer p.poolMutex.Unlock()
			for key, val := range p.pool {
				ch <- listenPoolTuple{key, val}
			}
			wg.Done()
		}()
		wg.Wait()
		close(ch)
	}()
	return ch
}
