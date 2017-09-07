package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	msgLengthMin int = 9
)

var (
	ErrMessageDataTooShort = errors.New("message data is too short")
)

type TMSG struct {
	Type      uint8
	TunnelID  uint32
	ChannelID uint32
	Payload   []byte
}

func (m *TMSG) String() string {
	return fmt.Sprintf("emsg(Type: %d TunnelID:%d ChannelID:%d L:%d)", m.Type, m.TunnelID, m.ChannelID, len(m.Payload))
}

func (m *TMSG) Len() int {
	return msgLengthMin + len(m.Payload)
}

func (m *TMSG) PayloadLen() int {
	return len(m.Payload)
}

func (m *TMSG) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, m.Type)
	binary.Write(buf, binary.LittleEndian, m.TunnelID)
	binary.Write(buf, binary.LittleEndian, m.ChannelID)

	if len(m.Payload) > 0 {
		binary.Write(buf, binary.LittleEndian, m.Payload)
	}

	return buf.Bytes()
}

func LoadTMSG(data []byte) (*TMSG, error) {
	if len(data) < msgLengthMin {
		return nil, ErrMessageDataTooShort
	}
	return &TMSG{
		Type:      data[0],
		TunnelID:  binary.LittleEndian.Uint32(data[1:5]),
		ChannelID: binary.LittleEndian.Uint32(data[5:9]),
		Payload:   data[9:],
	}, nil
}
