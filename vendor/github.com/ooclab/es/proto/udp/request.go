package udp

import "errors"

const (
	requestTypeQueryReceive uint8 = iota
)

var (
	errRequestUnknwonType = errors.New("unknown request type")
)
