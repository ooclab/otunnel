package udp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	numRetransmit  = 9
	defaultTimeout = 100
	maxTimeout     = 1600

	// TODO: sendWindowSize should changed dynamically
	defaultSendWindowSize = 64
	minSendWindowSize     = 8
	maxSendWindowSize     = 1024

	defaultConnTranSize   = 10
	defaultConnTimeout    = 30 * time.Second
	defaultPingInterval   = 6 * time.Second
	defaultPingTimeout    = 3 * time.Second
	defaultRequestTimeout = 12 * time.Second
	defaultSendingTimeout = 60 * time.Second // FIXME!

	maxRecvPoolSize = 10
	maxSendPoolSize = 10

	sendMsgMaxTimes = 999              // FIXME!
	maxMsgSize      = 1024 * 1024 * 16 // 16M

	// response status
	responseStatusUnknownType = 0
	queryReceiveNotExist      = 1
	queryReceiveCompleted     = 2
	queryReceiveNotCompleted  = 3
)

var (
	// ErrTimeout is commont timeout error
	ErrTimeout = errors.New("timeout")
	// ErrConnectionShutdown is a chan single of connection shutdown
	ErrConnectionShutdown = errors.New("connection is shutdown")
	// ErrSegTypeUnknown is the error abount unknown message type
	ErrSegTypeUnknown = errors.New("unknown message type")

	errSendingListFull = errors.New("sending list is full")
	errRecvingListFull = errors.New("recving list is full")

	errSegmentChecksum     = errors.New("segment checksum error")
	errClientExist         = errors.New("client is exist in ClientPool")
	errSegmentBodyTooLarge = errors.New("segment body is too large")

	errTransIDTooLarge = errors.New("transID is larger than defaultConnTranSize")
)

type msgRecving struct {
	readBuf        bytes.Buffer
	needLength     uint32
	readLength     uint32
	saved          map[uint16]*segment // saved seg
	nextID         uint16
	largestOrderID uint16 // the largest order id saved

	// !IMPORTANT! completed is a fag
	// It means this msgRecving should be take if re trans message incoming and this flag is true
	completed bool
	lock      sync.Mutex
}

func newMsgRecving() *msgRecving {
	return &msgRecving{
		saved: map[uint16]*segment{},
	}
}

func (m *msgRecving) GetMissing() (uint16, []uint16) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.completed {
		return 0, nil
	}

	ml := []uint16{}
	for i := m.nextID; i < m.largestOrderID; i++ {
		if _, ok := m.saved[i]; !ok {
			ml = append(ml, i)
		}
	}
	return m.largestOrderID, ml
}

// TODO: should return a number to determine missing segment grade, maybe oid-m.nextID ?
func (m *msgRecving) Save(seg *segment) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	oid := seg.h.OrderID()
	if oid < m.nextID || (oid >= m.nextID && m.saved[oid] != nil) {
		logrus.Warnf("dumplicate segment: %s", seg.h.String())
		return nil, nil
	}

	m.readLength += uint32(len(seg.b))
	if m.largestOrderID < oid {
		m.largestOrderID = oid
	}

	if oid == m.nextID {
		if oid == 0 {
			// FIXME!
			m.needLength = binary.BigEndian.Uint32(seg.b[0:4])
			m.readBuf.Write(seg.b[4:])
		} else {
			m.readBuf.Write(seg.b)
		}
		// clean current segment
		if _, ok := m.saved[oid]; ok {
			delete(m.saved, oid)
		}
		for {
			m.nextID++
			v, ok := m.saved[m.nextID]
			if !ok {
				break
			}
			m.readBuf.Write(v.b)
			delete(m.saved, m.nextID)
		}
	} else {
		m.saved[oid] = seg
	}

	// FIXME: readLength is enough?
	if m.needLength > 0 && m.needLength <= m.readLength {
		m.completed = true
		if len(m.saved) > 0 {
			// read message completed
			sl := []uint16{}
			for k := range m.saved {
				sl = append(sl, k)
			}
			sort.Sort(SIUInt16Slice(sl))
			for _, orderID := range sl {
				m.readBuf.Write(m.saved[orderID].b)
			}
		}
		// TODO: cleanup ?
		return m.readBuf.Bytes(), nil
	}

	return nil, nil
}

func (m *msgRecving) IsCompleted() bool {
	m.lock.Lock()
	b := m.completed
	m.lock.Unlock()
	return b
}

type msgSending struct {
	types    uint8
	flags    uint16
	streamID uint32
	transID  uint16
	message  []byte
}

func newMsgSending(types uint8, flags uint16, streamID uint32, transID uint16, message []byte) *msgSending {
	length := len(message)
	multiHdr := make([]byte, 4)
	binary.BigEndian.PutUint32(multiHdr, uint32(length+4))
	message = append(multiHdr, message...)

	return &msgSending{
		types:    types,
		flags:    flags,
		streamID: streamID,
		transID:  transID,
		message:  message,
	}
}

func (m *msgSending) segmentCount() uint16 {
	length := len(m.message)
	c := length / segmentBodyMaxSize
	if length%segmentBodyMaxSize != 0 {
		c++
	}
	return uint16(c)
}

func (m *msgSending) IterBufferd() <-chan *segment {
	length := len(m.message)
	sum := int(m.segmentCount())
	ch := make(chan *segment, sum)
	go func() {
		for i := 0; i < sum; i++ {
			end := (i + 1) * segmentBodyMaxSize
			if end > length {
				end = length
			}
			b := m.message[i*segmentBodyMaxSize : end]
			seg, _ := newSegment(m.types, m.flags, m.streamID, m.transID, uint16(i), b)
			ch <- seg
		}
		close(ch)
	}()
	return ch
}

func (m *msgSending) GetSegmentByOrderID(orderID uint16) *segment {
	start := int(orderID) * segmentBodyMaxSize
	end := start + segmentBodyMaxSize
	if end > len(m.message) {
		end = len(m.message)
	}
	b := m.message[start:end]
	seg, _ := newSegment(m.types, m.flags, m.streamID, m.transID, orderID, b)
	return seg
}

// Conn is a UDP implement of es.Conn
type Conn struct {
	c     *net.UDPConn
	raddr *net.UDPAddr
	id    uint32

	rl      []*msgRecving // recving list
	rlMutex sync.Mutex

	sl      []*msgSending // sending list
	slMutex sync.Mutex

	slWait      map[uint16]chan struct{} // wait transID
	slWaitMutex sync.Mutex

	// wait sending complete single
	ss      map[uint16]chan struct{}
	ssMutex sync.Mutex
	sws     int // the max segment in a sending loop

	lastActiveMutex sync.Mutex
	lastActive      time.Time

	inbound chan []byte

	// requests is used to send a inner request
	requests     map[uint32]chan []byte
	requestID    uint32
	requestMutex sync.Mutex

	// pings is used to track inflight pings
	pings    map[uint32]chan struct{}
	pingID   uint32
	pingLock sync.Mutex

	shutdownCh chan struct{}
}

func newConn(conn *net.UDPConn, raddr *net.UDPAddr, id uint32) *Conn {
	return &Conn{
		c:          conn,
		raddr:      raddr,
		id:         id,
		rl:         make([]*msgRecving, defaultConnTranSize),
		sl:         make([]*msgSending, defaultConnTranSize),
		ss:         make(map[uint16]chan struct{}),
		sws:        defaultSendWindowSize,
		lastActive: time.Now(),
		inbound:    make(chan []byte, 1),

		pings:    make(map[uint32]chan struct{}),
		requests: make(map[uint32]chan []byte),
		slWait:   make(map[uint16]chan struct{}),

		shutdownCh: make(chan struct{}),
	}
}

// RemoteAddr get the address of remote endpoint
func (c *Conn) RemoteAddr() net.Addr {
	return c.raddr
}

// LocalAddr get the address of local endpoint
func (c *Conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *Conn) String() string {
	return fmt.Sprintf("conn %d: %s(L) -- %s(R)", c.id, c.LocalAddr(), c.RemoteAddr())
}

func (c *Conn) getRecving(transID uint16) (*msgRecving, error) {
	if transID >= defaultConnTranSize {
		return nil, errTransIDTooLarge
	}
	c.rlMutex.Lock()
	recving := c.rl[transID]
	c.rlMutex.Unlock()
	return recving, nil
}

func (c *Conn) setRecving(transID uint16, recving *msgRecving) error {
	if transID >= defaultConnTranSize {
		return errTransIDTooLarge
	}
	c.rlMutex.Lock()
	c.rl[transID] = recving
	c.rlMutex.Unlock()
	return nil
}

func (c *Conn) getLastActive() time.Time {
	c.lastActiveMutex.Lock()
	lt := c.lastActive
	c.lastActiveMutex.Unlock()
	return lt
}

func (c *Conn) handle(msg []byte) error {
	c.lastActiveMutex.Lock()
	c.lastActive = time.Now()
	c.lastActiveMutex.Unlock()

	seg, err := loadSegment(msg)
	if err != nil {
		logrus.Errorf("load segment failed: %s", err)
		return err
	}

	types := seg.h.Type()

	switch types {
	case segTypeMsgSYN:
		err = c.handlePingSYN(seg)
	case segTypeMsgPingReq:
		err = c.handlePingReq(seg)
	case segTypeMsgPingRep:
		err = c.handlePingRep(seg)
	case segTypeMsgReq:
		err = c.handleReq(seg)
	case segTypeMsgRep:
		err = c.handleRep(seg)
	case segTypeMsgReceived:
		err = c.handleReceived(seg)
	case segTypeMsgReTrans:
		err = c.handleReTrans(seg)
	case segTypeMsgTrans:
		err = c.handleTrans(seg)
	default:
		err = c.handleUnknown(seg)
	}

	return err
}

func (c *Conn) handlePingSYN(seg *segment) error {
	seg = newACKSegment(seg.b) // FIXME!
	return c.write(seg.bytes())
}

func (c *Conn) handlePingReq(seg *segment) error {
	seg = newPingRepSegment(c.id, seg.b)
	return c.write(seg.bytes())
}

func (c *Conn) handlePingRep(seg *segment) error {
	// notice ping wait
	pingID := binary.BigEndian.Uint32(seg.b[0:4])
	c.pingLock.Lock()
	ch := c.pings[pingID]
	if ch != nil {
		delete(c.pings, pingID)
		close(ch)
	}
	c.pingLock.Unlock()
	return nil
}

func (c *Conn) handleReq(seg *segment) error {
	if len(seg.b) < 5 {
		return errors.New("invalid request messgae")
	}
	types := seg.b[4] // FIXME!
	switch types {
	case requestTypeQueryReceive:
		return c.handleReqQueryReceive(seg)
	default:
		logrus.Errorf("unknown request types: %d", types)
		seg, _ = newSegment(segTypeMsgRep, 0, seg.h.StreamID(), 0, 0, []byte{responseStatusUnknownType})
		c.write(seg.bytes())
		return errRequestUnknwonType
	}
}

// handleReqQueryReceive query recving status of the specified msg
func (c *Conn) handleReqQueryReceive(seg *segment) error {
	transID := seg.h.TransID()
	recving, err := c.getRecving(transID)
	if err != nil {
		return err
	}
	if recving == nil {
		return c.responseQueryReceive(seg, queryReceiveNotExist)
	}
	if recving.IsCompleted() {
		return c.responseQueryReceive(seg, queryReceiveCompleted)
	}

	// not completed
	largestOrderID, missingOrderIDList := recving.GetMissing()

	// !IMPORTANT! segment size limit!
	max := len(missingOrderIDList)
	if max > (segmentBodyMaxSize-7)/2 {
		max = (segmentBodyMaxSize - 7) / 2
	}
	if max > c.sws {
		max = c.sws // FIXME! test
	}

	b := make([]byte, 7+max*2)
	copy(b[0:4], seg.b[0:4])
	b[4] = queryReceiveNotCompleted
	binary.BigEndian.PutUint16(b[5:7], largestOrderID)
	for i := 0; i < max; i++ {
		binary.BigEndian.PutUint16(b[7+i*2:7+i*2+2], missingOrderIDList[i])
	}
	seg, _ = newSegment(segTypeMsgRep, 0, c.id, transID, 0, b)
	return c.write(seg.bytes())
}

func (c *Conn) responseQueryReceive(seg *segment, status uint8) error {
	b := make([]byte, 5)
	copy(b[0:4], seg.b[0:4])
	b[4] = status
	seg, _ = newSegment(segTypeMsgRep, 0, c.id, seg.h.TransID(), 0, b)
	return c.write(seg.bytes())
}

func (c *Conn) handleRep(seg *segment) error {
	// notice ping wait
	requestID := binary.BigEndian.Uint32(seg.b[0:4])
	c.requestMutex.Lock()
	ch := c.requests[requestID]
	if ch != nil {
		// add timeout!
		ch <- seg.b[4:]
		delete(c.requests, requestID)
		close(ch)
	}
	c.requestMutex.Unlock()
	return nil
}

func (c *Conn) handleReceived(seg *segment) error {
	// FIXME!
	transID := seg.h.TransID()
	c.slWaitMutex.Lock()
	ch := c.slWait[transID]
	if ch != nil {
		delete(c.slWait, transID)
		close(ch)
	}
	c.slWaitMutex.Unlock()
	return nil
}

func (c *Conn) handleReTrans(seg *segment) error {
	logrus.Errorf("re trans message have not completed!")
	return nil
}

func (c *Conn) handleTrans(seg *segment) error {
	transID := seg.h.TransID()
	recving, err := c.getRecving(transID)
	if err != nil {
		return err
	}
	// !IMPORTANT! recving.completed
	if recving == nil || recving.IsCompleted() {
		recving = newMsgRecving()
		c.setRecving(transID, recving)
	}
	// fmt.Printf("%p recving: nextID = %d, transID = %d, orderID = %d, %s\n", recving, recving.nextID, transID, seg.h.OrderID(), hex.EncodeToString(seg.h.Checksum()[:]))
	msg, err := recving.Save(seg)
	if err != nil {
		return err
	}
	if msg != nil {
		go func() { c.inbound <- msg }() // FIXME!! do not lock at here!
		// send msg received
		seg, _ := newSegment(segTypeMsgReceived, 0, c.id, transID, 0, nil)
		return c.write(seg.bytes())
	}
	// TODO: recving.Save() should return a grade for change remote sws(Sending Window Size)!
	return nil
}

func (c *Conn) handleUnknown(seg *segment) error {
	logrus.Errorf("unknown type segment: %s", seg.h.String())
	return ErrSegTypeUnknown
}

// RecvMsg recv a single message
func (c *Conn) RecvMsg() ([]byte, error) {
	// TODO: timeout
	// Wait for a response
	select {
	case msg := <-c.inbound:
		return msg, nil
	case <-c.shutdownCh:
		return nil, ErrConnectionShutdown
	}
}

// SendMsg send a single message
func (c *Conn) SendMsg(message []byte) error {
	length := len(message)
	if length <= 0 {
		return errors.New("empty message")
	}

	// TODO: timeout
	// get sending stor
	var sending *msgSending
	for {
		c.slMutex.Lock()
		for i, v := range c.sl {
			if v == nil {
				sending = newMsgSending(segTypeMsgTrans, 0, c.id, uint16(i), message)
				c.sl[i] = sending
				defer func() { c.sl[i] = nil }() // FIXME!
				break
			}
		}
		c.slMutex.Unlock()
		if sending != nil {
			break
		}
		fmt.Println("wait transID")
		time.Sleep(100 * time.Millisecond)
	}

	ch := make(chan struct{})
	c.slWaitMutex.Lock()
	c.slWait[sending.transID] = ch
	c.slWaitMutex.Unlock()
	var remain int

	for i := 0; i < sendMsgMaxTimes; {
	QUERY:
		i++
		remain = c.sws
		if i > 1 {
			// must query remote endpoint before send message again
			status, largestOrderID, missing, err := c.queryMsgReceive(sending)
			if err != nil {
				return err // FIXME!
			}
			if status == queryReceiveCompleted {
				return nil
			}
			if status == queryReceiveNotCompleted {
				maxOrderID := sending.segmentCount() - 1
				// handle missing
				for _, orderID := range missing {
					if remain <= 0 {
						goto QUERY
					}
					if orderID > maxOrderID {
						logrus.Error("SHOULD NOT: seg is null: ", orderID, len(sending.message))
						return errors.New("orderID is too large")
					}
					seg := sending.GetSegmentByOrderID(orderID)
					c.write(seg.bytes())
					remain--
				}
				// handle largestOrderID
				for orderID := largestOrderID + 1; orderID <= maxOrderID; orderID++ {
					if remain <= 0 {
						goto QUERY
					}
					seg := sending.GetSegmentByOrderID(orderID)
					c.write(seg.bytes())
					remain--
				}
				goto WAIT
			}
		}

		// sending full message
		for seg := range sending.IterBufferd() {
			if remain <= 0 {
				goto QUERY
			}
			if err := c.write(seg.bytes()); err != nil {
				return err
			}
			remain--
		}

	WAIT:
		// waiting message received success
		select {
		case <-ch:
			return nil
		case <-time.After(defaultSendingTimeout):
		case <-c.shutdownCh:
			return ErrConnectionShutdown
		}
	}

	// clean
	c.slWaitMutex.Lock()
	delete(c.slWait, sending.transID)
	c.slWaitMutex.Unlock()

	return ErrTimeout
}

func (c *Conn) Read(p []byte) (n int, err error) {
	msg, err := c.RecvMsg()
	if err == nil {
		copy(p, msg)
	}
	return len(msg), err
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if len(p) <= maxMsgSize {
		err = c.SendMsg(p)
		if err != nil {
			return
		}
		return len(p), nil
	}

	total := len(p)
	var start, end, read int
	for start < total {
		end += maxMsgSize
		if end >= total {
			end = total
		}
		err = c.SendMsg(p[start:end])
		if err != nil {
			return read, err
		}

		n += (end - start)
		start = end
	}
	return read, nil
}

func (c *Conn) queryMsgReceive(s *msgSending) (status uint8, largestOrderID uint16, missing []uint16, err error) {
	id, ch := c.genRequestIDChan()
	b := make([]byte, 5)
	binary.BigEndian.PutUint32(b[0:4], id)
	b[4] = requestTypeQueryReceive
	seg, _ := newSegment(segTypeMsgReq, s.flags, c.id, s.transID, 0, b)

	for i := 0; i < 999; i++ {
		if err = c.write(seg.bytes()); err != nil {
			logrus.Errorf("queryMsgReceive: write segment failed: %s", err)
			return
		}

		// Wait for a response
		select {
		case res := <-ch:
			status = res[0]
			if status == queryReceiveCompleted || status == queryReceiveNotExist {
				return
			}
			// not completed
			if len(res) < 3 {
				err = errors.New("no completed need more than 3 bytes")
				return
			}
			largestOrderID = binary.BigEndian.Uint16(res[1:3])
			missing = []uint16{}
			for j := 0; j < (len(res)-3)/2; j++ {
				orderID := binary.BigEndian.Uint16(res[3+j*2 : 3+j*2+2])
				missing = append(missing, orderID)
			}
			return // success
		case <-time.After(1000 * time.Millisecond):
			continue // retry
			// FIXME! just one time.After
		case <-time.After(defaultRequestTimeout):
			c.requestMutex.Lock()
			delete(c.requests, id)
			c.requestMutex.Unlock()
			close(ch)
			err = ErrTimeout
			return
		case <-c.shutdownCh:
			err = ErrConnectionShutdown
			return
		}
	}

	err = errors.New("query try many times")
	return
}

func (c *Conn) write(b []byte) error {
	_, err := c.c.WriteToUDP(b, c.raddr)
	return err
}

// Ping is used to measure the RTT response time
func (c *Conn) Ping() (time.Duration, error) {
	ch := make(chan struct{})

	// Get a new ping id, mark as pending
	c.pingLock.Lock()
	id := c.pingID
	c.pingID++
	c.pings[id] = ch
	c.pingLock.Unlock()

	// Send the ping request
	seg := newPingReqSegment(c.id, id)
	c.c.WriteToUDP(seg.bytes(), c.raddr)

	// Wait for a response
	start := time.Now()
	select {
	case <-ch:
	case <-time.After(defaultPingTimeout):
		c.pingLock.Lock()
		delete(c.pings, id)
		c.pingLock.Unlock()
		return 0, ErrTimeout
	case <-c.shutdownCh:
		return 0, ErrConnectionShutdown
	}

	// TODO: compute time duration
	return time.Now().Sub(start), nil
}

func (c *Conn) genRequestIDChan() (id uint32, ch chan []byte) {
	ch = make(chan []byte)

	// Get a new request id, mark as pending
	c.requestMutex.Lock()
	for {
		c.requestID++
		if _, ok := c.requests[c.requestID]; !ok {
			break
		}
	}
	id = c.requestID
	c.requests[id] = ch
	c.requestMutex.Unlock()
	return id, ch
}

// request send a request to remote endpoint, and wait the response
func (c *Conn) request(msg []byte) ([]byte, error) {
	id, ch := c.genRequestIDChan()

	// Send the request
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint32(hdr, id)
	msg = append(hdr, msg...)

	seg := newReqSegment(c.id, msg)
	c.c.WriteToUDP(seg.bytes(), c.raddr)

	// Wait for a response
	select {
	case rmsg := <-ch:
		return rmsg, nil
	case <-time.After(defaultRequestTimeout):
		c.requestMutex.Lock()
		delete(c.requests, id)
		c.requestMutex.Unlock()
		return nil, ErrTimeout
	case <-c.shutdownCh:
		return nil, ErrConnectionShutdown
	}
}

// Close close this connection
func (c *Conn) Close() error {
	// FIXME: close is not completed
	close(c.shutdownCh)
	// close(c.inbound)
	return nil
}

type clientPool struct {
	nextClientID uint32
	idAddrMap    map[uint32]string
	m            *sync.Mutex
}

func newClientPool() *clientPool {
	return &clientPool{
		nextClientID: 1,
		idAddrMap:    map[uint32]string{},
		m:            &sync.Mutex{},
	}
}

func (p *clientPool) newClientID() (id uint32) {
	p.m.Lock()
	for {
		id = p.nextClientID
		p.nextClientID++
		if _, ok := p.idAddrMap[id]; !ok {
			break
		}
	}
	p.m.Unlock()
	return
}

// connPool manage all connections
type connPool struct {
	addrConnMap map[string]*Conn
	m           *sync.Mutex
}

func newConnPool() *connPool {
	return &connPool{
		addrConnMap: map[string]*Conn{},
		m:           &sync.Mutex{},
	}
}

// Get get the connection specified by address string
func (p *connPool) Get(addr net.Addr) (*Conn, bool) {
	p.m.Lock()
	c, ok := p.addrConnMap[addr.String()]
	p.m.Unlock()
	return c, ok
}

// New create a special single connection
func (p *connPool) New(conn *net.UDPConn, raddr *net.UDPAddr, id uint32) (*Conn, error) {
	addr := raddr.String()
	p.m.Lock()
	_, ok := p.addrConnMap[addr]
	p.m.Unlock()
	if ok {
		return nil, errClientExist
	}
	c := newConn(conn, raddr, id)
	p.m.Lock()
	p.addrConnMap[addr] = c
	p.m.Unlock()
	return c, nil
}

// Delete remove a conn from client pool
func (p *connPool) Delete(conn *Conn) error {
	p.m.Lock()
	defer p.m.Unlock()
	addr := conn.raddr.String()
	if _, ok := p.addrConnMap[addr]; !ok {
		return errors.New("delete: no such addr in addrConnMap")
	}
	delete(p.addrConnMap, addr)
	return nil
}

// GarbageCollection delete the disconnected clients
func (p *connPool) GarbageCollection() {
	addrs := []string{}
	p.m.Lock()
	for addr, conn := range p.addrConnMap {
		if time.Since(conn.getLastActive()) > defaultConnTimeout {
			addrs = append(addrs, addr)
		}
	}
	p.m.Unlock()

	p.m.Lock()
	for _, addr := range addrs {
		p.addrConnMap[addr].Close() // FIXME!
		delete(p.addrConnMap, addr)
		logrus.Debugf("client %s is timeout, delete it", addr)
	}
	p.m.Unlock()
}

type udpserver struct {
	c *net.UDPConn

	clients  *clientPool
	connPool *connPool

	clientCh chan *Conn

	closed bool
}

func (p *udpserver) garbageCollection() {
	for {
		start := time.Now()
		p.connPool.GarbageCollection()
		time.Sleep(10*time.Second - time.Now().Sub(start))
	}
}

func (p *udpserver) recv() error {
	// FIXME!
	go p.garbageCollection()

	buf := make([]byte, segmentMaxSize)
	for {
		n, raddr, err := p.c.ReadFromUDP(buf)
		if err != nil {
			if p.closed {
				return nil
			}
			// FIXME!
			if strings.Contains(err.Error(), "use of closed network connection") {
				logrus.Debugf("conn is closed, quit recv()")
				p.closed = true
				return nil
			}
			logrus.Errorf("ReadFromUDP error: %s", err)
			return err
		}

		conn, ok := p.connPool.Get(raddr)
		if !ok {
			// save new client
			id := p.clients.newClientID()
			conn, err = p.connPool.New(p.c, raddr, id)
			if err != nil {
				logrus.Errorf("save new client failed: %s", err)
				// TODO: notice schema
				seg := newACKSegment([]byte("error: create client conn"))
				p.c.WriteToUDP(seg.bytes(), raddr)
				continue
			}
			p.clientCh <- conn
		}

		// handle in
		if err := conn.handle(buf[0:n]); err != nil {
			logrus.Errorf("handle msg(from %s) failed: %s", raddr.String(), err)
		}
	}
}

// Accept wait the new client connection incoming
func (p *udpserver) Accept() (*Conn, error) {
	return <-p.clientCh, nil
}

// TODO: add lock
func (p *udpserver) Close() error {
	if p.closed {
		return nil
	}

	p.closed = true
	return nil
}

// ClientSocket is a UDP implement of Socket
type ClientSocket struct {
	udpserver
	raddr *net.UDPAddr
}

// NewClientSocket create a client socket
func NewClientSocket(conn *net.UDPConn, raddr *net.UDPAddr) (*ClientSocket, *Conn, error) {
	sock := &ClientSocket{
		udpserver: udpserver{
			c:        conn,
			clients:  newClientPool(),
			connPool: newConnPool(),
			clientCh: make(chan *Conn, 1),
		},
		raddr: raddr,
	}
	c, err := sock.handshake() // FIXME! quit?
	if err != nil {
		return nil, nil, err
	}
	go sock.pingLoop(c)
	go sock.recv()
	return sock, c, nil
}

func (p *ClientSocket) handshake() (*Conn, error) {
	for {
		if conn, err := p._handshake(); err == nil {
			return conn, err
		}
		time.Sleep(6 * time.Second)
	}
}

func (p *ClientSocket) _handshake() (*Conn, error) {
	// send heartbeat and wait
	seg := newSYNSegment()
	_, err := p.c.WriteToUDP(seg.bytes(), p.raddr)
	if err != nil {
		logrus.Warnf("handshake: write segment failed: %s", err)
		return nil, err
	}

	buf := make([]byte, segmentBodyMaxSize)

	// read
	n, raddr, err := p.c.ReadFromUDP(buf)
	if raddr.String() != p.raddr.String() {
		logrus.Debugf("p.raddr.String() = %s, raddr.String() = %s", p.raddr.String(), raddr.String())
		logrus.Warnf("unknown from addr: %s", raddr.String())
	}
	if err != nil {
		logrus.Warnf("handshake: read segment failed: %s", err)
		return nil, err
	}

	seg, err = loadSegment(buf[0:n])
	if err != nil {
		logrus.Warnf("handshake: loadSegment failed: %s", err)
		return nil, err
	}
	if seg.h.Type() != segTypeMsgACK {
		logrus.Warnf("handshake: segment type is %d, not segTypeMsgSYN(%d)", seg.h.Type(), segTypeMsgSYN)
		return nil, errors.New("segment type is not segTypeMsgACK")
	}
	if string(seg.b) != handshakeKey {
		logrus.Warnf("handshake: response segment body is mismatch")
		return nil, errors.New("response segment body is mismatch")
	}

	// TODO: check streamID
	return p.connPool.New(p.c, p.raddr, seg.h.StreamID())
}

func (p *ClientSocket) pingLoop(c *Conn) {
	for {
		c.Ping() // ping timeout
		time.Sleep(defaultPingInterval)
	}
}

// ServerSocket is a UDP implement of socket
type ServerSocket struct {
	udpserver
}

// NewServerSocket create a UDPConn
func NewServerSocket(conn *net.UDPConn) (*ServerSocket, error) {
	sock := &ServerSocket{
		udpserver: udpserver{
			c:        conn,
			clients:  newClientPool(),
			connPool: newConnPool(),
			clientCh: make(chan *Conn, 1),
		},
	}
	go sock.recv()
	return sock, nil
}
