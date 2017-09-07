package session

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	eMSGLengthMin int = 5
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

func (m *EMSG) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, m.Type)
	binary.Write(buf, binary.BigEndian, m.ID)

	if len(m.Payload) > 0 {
		binary.Write(buf, binary.BigEndian, m.Payload)
	}

	return buf.Bytes()
}

func LoadEMSG(data []byte) (*EMSG, error) {
	if len(data) < eMSGLengthMin {
		return nil, ErrMessageDataTooShort
	}
	return &EMSG{
		Type:    data[0],
		ID:      binary.BigEndian.Uint32(data[1:5]),
		Payload: data[5:],
	}, nil
}
