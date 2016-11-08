package emsg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	eMSGLengthMin int = 5
)

const (
	MSG_TYPE_REQ = 1
	MSG_TYPE_REP = 2
)

var (
	ErrMessageDataTooShort = errors.New("message data is too short")
)

type EMSG struct {
	Type    uint8
	ID      uint32
	Payload []byte
}

func (m *EMSG) String() string {
	return fmt.Sprintf("emsg(Type: %d ID:%d L:%d)", m.Type, m.ID, len(m.Payload))
}

func (m *EMSG) Len() int {
	return 1 + 4 + len(m.Payload)
}

func (m *EMSG) PayloadLen() int {
	return len(m.Payload)
}

func (m *EMSG) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m.Type)
	binary.Write(buf, binary.LittleEndian, m.ID)

	if len(m.Payload) > 0 {
		binary.Write(buf, binary.LittleEndian, m.Payload)
	}

	return buf.Bytes()
}

func LoadEMSG(data []byte) (*EMSG, error) {
	if len(data) < eMSGLengthMin {
		return nil, ErrMessageDataTooShort
	}
	return &EMSG{
		Type:    data[0],
		ID:      binary.LittleEndian.Uint32(data[1:5]),
		Payload: data[5:],
	}, nil
}
