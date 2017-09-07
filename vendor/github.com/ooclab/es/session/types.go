package session

const (
	MsgTypeRequest  uint8 = 1
	MsgTypeResponse uint8 = 2
	MsgTypeClose    uint8 = 3 // TODO: for close session
)

type Request struct {
	Action string
	Body   []byte
}

type Response struct {
	Status string
	Body   []byte
}
