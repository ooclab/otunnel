package link

import (
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/ooclab/es"
	"github.com/ooclab/es/session"
	"github.com/ooclab/es/tunnel"
	"github.com/sirupsen/logrus"
)

// Define error
var (
	ErrLinkShutdown     = errors.New("link is shutdown")
	ErrTimeout          = errors.New("timeout")
	ErrKeepaliveTimeout = errors.New("keepalive error")
	ErrMsgPingInvalid   = errors.New("invalid ping message")
)

const (
	sizeOfType   = 1
	sizeOfLength = 3
	headerSize   = sizeOfType + sizeOfLength

	defaultInterval = 6 * time.Second  // min
	maxLinkIdle     = 60 * time.Second // TODO: needed ?
)

// LinkConfig reserved for config
type LinkConfig struct {
	// ID need to be started differently
	IsServerSide bool

	// KeepaliveInterval is how often to perform the keep alive
	KeepaliveInterval time.Duration

	// ConnectionWriteTimeout is meant to be a "safety valve" timeout after
	// we which will suspect a problem with the underlying connection and
	// close it. This is only applied to writes, where's there's generally
	// an expectation that things will move along quickly.
	ConnectionWriteTimeout time.Duration
}

// Link is the main connection between two endpoint
type Link struct {
	ID     uint32
	config *LinkConfig
	log    *logrus.Entry

	sessionManager *session.Manager
	tunnelManager  *tunnel.Manager

	outbound chan []byte

	// for underlying loop control
	stopCh   chan struct{}
	stopLock sync.Mutex
	wg       *sync.WaitGroup

	lastRecvTime      time.Time
	lastRecvTimeMutex *sync.Mutex

	// pings is used to track inflight pings
	pings    map[uint32]chan struct{}
	pingID   uint32
	pingLock sync.Mutex

	shutdownCh   chan struct{}
	shutdownLock sync.Mutex

	defaultOpenTunnel OpenTunnelFunc
}

// NewLink create a new link
func NewLink(config *LinkConfig) *Link {
	return newLink(config, nil)
}

// NewLinkCustom create a new link by custom session.RequestHandler
func NewLinkCustom(config *LinkConfig, hdr session.RequestHandler) *Link {
	return newLink(config, hdr)
}

func newLink(config *LinkConfig, hdr session.RequestHandler) *Link {
	logrus.Debugf("prepare to create new link with config %#v", config)
	if config == nil {
		config = &LinkConfig{}
	}
	if config.KeepaliveInterval == 0 {
		config.KeepaliveInterval = 30 * time.Second
	}
	if config.ConnectionWriteTimeout == 0 {
		config.ConnectionWriteTimeout = 10 * time.Second
	}
	l := &Link{
		config:            config,
		outbound:          make(chan []byte, 1),
		lastRecvTimeMutex: &sync.Mutex{},

		pings:      make(map[uint32]chan struct{}),
		shutdownCh: make(chan struct{}),
	}
	l.log = logrus.WithFields(logrus.Fields{
		"from": "link",
		"id":   l.ID,
	})
	l.sessionManager = session.NewManager(config.IsServerSide, l.outbound)
	l.tunnelManager = tunnel.NewManager(config.IsServerSide, l.outbound, l.sessionManager)
	if hdr == nil {
		hdr = newRequestHandler([]session.Route{
			{"/tunnel", defaultTunnelCreateHandler(l.tunnelManager)},
		})
	}
	l.sessionManager.SetRequestHandler(hdr)
	// TODO: custom defaultOpenTunnel func
	l.defaultOpenTunnel = defaultOpenTunnel(l.sessionManager, l.tunnelManager)

	// run keepalive
	go func() {
		if err := l.keepalive(); err != nil {
			l.log.Errorf("link %d: keepalive failed: %s", l.ID, err)
		}
		l.log.Debugf("link %d: stop keepalive", l.ID)
	}()

	l.log.Debug("create link success")
	return l
}

// IsClosed does a safe check to see if we have shutdown
func (l *Link) IsClosed() bool {
	select {
	case <-l.shutdownCh:
		return true
	default:
		return false
	}
}

// Close is used to close the link
func (l *Link) Close() error {
	if l.IsClosed() {
		l.log.Warn("link is closed already")
		return nil
	}

	l.shutdownLock.Lock()
	close(l.shutdownCh)
	l.shutdownLock.Unlock()

	close(l.outbound)
	// TODO: close sessions & tunnles
	l.tunnelManager.Close()
	l.sessionManager.Close()
	return nil
}

// IsStopped does a safe check to see if we have shutdown
func (l *Link) IsStopped() bool {
	select {
	case <-l.stopCh:
		return true
	default:
		return false
	}
}

// Stop close the current transaction underlying conn
func (l *Link) Stop() error {
	if l.IsStopped() {
		l.log.Warn("link is stopped already")
		return nil
	}

	l.stopLock.Lock()
	close(l.stopCh)
	l.stopLock.Unlock()

	return nil
}

// Ping is used to measure the RTT response time
func (l *Link) Ping() (time.Duration, error) {
	// Get a channel for the ping
	ch := make(chan struct{})

	// Get a new ping id, mark as pending
	l.pingLock.Lock()
	id := l.pingID
	l.pingID++
	l.pings[id] = ch
	l.pingLock.Unlock()

	// Send the ping request
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, id)
	l.outbound <- append([]byte{es.LinkMsgTypePingRequest}, payload...)

	// Wait for a response
	start := time.Now()
	select {
	case <-ch:
		// Compute the RTT
		return time.Now().Sub(start), nil
	case <-time.After(l.config.ConnectionWriteTimeout):
		l.pingLock.Lock()
		delete(l.pings, id) // Ignore it if a response comes later.
		l.pingLock.Unlock()
		return 0, ErrTimeout
	case <-l.shutdownCh:
		return 0, ErrLinkShutdown
	}
}

// keepalive is a long running goroutine that periodically does
// a ping to keep the connection alive.
func (l *Link) keepalive() error {
	l.log.WithFields(logrus.Fields{
		"interval":    l.config.KeepaliveInterval,
		"max_timeout": l.config.ConnectionWriteTimeout,
	}).Debug("start keepalive")
	interval := defaultInterval
	for {
		select {
		case <-time.After(interval):
			idle := l.recvIdlenessTime()
			if idle > l.config.KeepaliveInterval {
				interval = defaultInterval
				rtt, err := l.Ping()
				if err != nil {
					l.log.WithFields(logrus.Fields{
						"error": err,
						"idle":  idle,
					}).Warn("ping failed")
					if idle > maxLinkIdle {
						l.log.WithFields(logrus.Fields{
							"idle": idle,
						}).Error("max link idle reach, close the underlying net.Conn")
						l.Stop() // notice the underlying loop
					}
					continue
				}
				l.log.WithField("rtt", rtt).Debug("ping success")
			} else {
				interval = l.config.KeepaliveInterval - idle
			}
		case <-l.shutdownCh:
			return nil
		}
	}
}

func (l *Link) recv(conn es.Conn) error {
	l.log.Debug("start underlying recv")
	for {
		m, err := conn.Recv()
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "closed") && !strings.Contains(err.Error(), "reset by peer") {
				l.log.WithFields(logrus.Fields{
					"error": err,
				}).Error("read failed")
			}
			// TODO: nil
			return err
		}

		l.updateLastRecvTime()

		mType, mData := m[0], m[1:]

		// dispatch
		switch mType {
		case es.LinkMsgTypeSession:
			err = l.sessionManager.HandleIn(mData)
		case es.LinkMsgTypeTunnel:
			err = l.tunnelManager.HandleIn(mData)
		case es.LinkMsgTypePingRequest:
			l.outbound <- append([]byte{es.LinkMsgTypePingResponse}, mData...)
		case es.LinkMsgTypePingResponse:
			err = l.handlePing(mData)
		default:
			l.log.WithField("type", mType).Error("unknown message type")
			// TODO:
			return errors.New("unknown message type")
		}

		if err != nil {
			return err
		}
	}
}

func (l *Link) send(conn es.Conn) error {
	l.log.Debug("start underlying send")
	for {
		select {
		case m := <-l.outbound:
			// FIXME!
			if m == nil {
				return errors.New("get nil from l.outbound")
			}
			err := conn.Send(m)
			if err != nil {
				l.log.WithField("error", err).Error("write data to conn failed")
				return err
			}
		case <-l.stopCh:
			l.log.Debug("got stop event, quit Link.send")
			return nil
		case <-l.shutdownCh:
			l.log.Debug("got shutdown event, quit Link.send")
			return nil
		}
	}
}

// Bind bind link with a underlying connection (tcp)
func (l *Link) Bind(conn es.Conn) error {
	l.wg = &sync.WaitGroup{}
	l.stopCh = make(chan struct{}, 1)

	go func() {
		l.wg.Add(1)
		if err := l.recv(conn); err != nil {
			l.log.WithField("error", err).Error("Link.recv quit")
		}
		// TODO: notice send
		l.Stop()
		l.wg.Done()
	}()
	go func() {
		l.wg.Add(1)
		if err := l.send(conn); err != nil {
			l.log.WithField("error", err).Error("Link.send quit")
		}
		// TODO: notice recv
		conn.Close()
		l.wg.Done()
	}()

	l.Ping() // TODO: wait ping success
	return nil
}

func (l *Link) Wait() {
	l.wg.Wait()
	l.wg = nil
	l.Stop()
	l.stopCh = nil
	l.log.Debug("wait completed")
}

// handlePing is invokde for a LinkMsgTypePing frame
func (l *Link) handlePing(payload []byte) error {
	if len(payload) != 4 {
		return ErrMsgPingInvalid
	}
	pingID := binary.LittleEndian.Uint32(payload)
	l.pingLock.Lock()
	ch := l.pings[pingID]
	if ch != nil {
		delete(l.pings, pingID)
		close(ch)
	}
	l.pingLock.Unlock()
	return nil
}

func (l *Link) NewSession() (*session.Session, error) {
	return l.sessionManager.New()
}

func (l *Link) updateLastRecvTime() {
	l.lastRecvTimeMutex.Lock()
	l.lastRecvTime = time.Now()
	l.lastRecvTimeMutex.Unlock()
}

func (l *Link) recvIdlenessTime() time.Duration {
	l.lastRecvTimeMutex.Lock()
	d := time.Since(l.lastRecvTime)
	l.lastRecvTimeMutex.Unlock()
	return d
}

// OpenTunnel open a tunnel
func (l *Link) OpenTunnel(proto string, localHost string, localPort int, remoteHost string, remotePort int, reverse bool) error {
	return l.defaultOpenTunnel(proto, localHost, localPort, remoteHost, remotePort, reverse)
}
