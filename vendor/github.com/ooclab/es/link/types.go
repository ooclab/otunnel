package link

// may be other default request func

// OpenTunnelFunc define a func about open tunnel
type OpenTunnelFunc func(proto string, localHost string, localPort int, remoteHost string, remotePort int, reverse bool) error

type tunnelCreateBody struct {
	ID uint32
}
