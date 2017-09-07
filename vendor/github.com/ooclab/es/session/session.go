package session

import (
	"encoding/json"

	"github.com/ooclab/es"
)

type Session struct {
	ID       uint32
	inbound  chan []byte
	outbound chan []byte
}

func newSession(id uint32, outbound chan []byte) *Session {
	return &Session{
		ID:       id,
		inbound:  make(chan []byte, 1),
		outbound: outbound,
	}
}

func (session *Session) Close() {
	close(session.inbound)
}

func (session *Session) HandleResponse(payload []byte) error {
	// logrus.Debugf("inner session : got response : %s", string(payload))
	session.inbound <- payload
	return nil
}

func (session *Session) sendAndWait(payload []byte) (respPayload []byte, err error) {
	// TODO:

	m := &EMSG{
		Type:    MsgTypeRequest,
		ID:      session.ID,
		Payload: payload,
	}
	session.outbound <- append([]byte{es.LinkMsgTypeSession}, m.Bytes()...)

	// TODO: timeout
	respPayload = <-session.inbound
	return respPayload, nil
}

func (session *Session) SendAndWait(r *Request) (resp *Response, err error) {
	reqData, err := json.Marshal(r)
	if err != nil {
		return
	}
	respData, err := session.sendAndWait(reqData)
	if err != nil {
		return
	}
	resp = &Response{}
	err = json.Unmarshal(respData, &resp)
	return
}

func (session *Session) SendJSONAndWait(request interface{}, response interface{}) error {
	reqData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	respData, err := session.sendAndWait(reqData)
	if err != nil {
		return err
	}
	return json.Unmarshal(respData, &response)
}
