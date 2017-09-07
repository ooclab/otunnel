package udp

// SIUInt16Slice define a uint16 slice sort interface
type SIUInt16Slice []uint16

func (c SIUInt16Slice) Len() int {
	return len(c)
}
func (c SIUInt16Slice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c SIUInt16Slice) Less(i, j int) bool {
	return c[i] < c[j]
}
