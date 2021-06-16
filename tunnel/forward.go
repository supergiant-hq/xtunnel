package tunnel

import (
	"net"
	"sync"

	p2pc "github.com/supergiant-hq/xnet/p2p/client"
)

type Forward struct {
	FromAddr    string
	ToAddr      string
	Conn        *p2pc.Connection
	Forwards    *sync.Map
	TCPListener *net.TCPListener
	UDPListener *net.UDPConn
	Closed      bool
	mutex       sync.Mutex
}

func (fwd *Forward) Close(reason string) {
	fwd.mutex.Lock()
	defer fwd.mutex.Unlock()

	if fwd.Closed {
		return
	}

	fwd.Conn.Close(reason)
	if fwd.TCPListener != nil {
		fwd.TCPListener.Close()
	}
	if fwd.UDPListener != nil {
		fwd.UDPListener.Close()
	}
	fwd.Forwards.Delete(fwd.FromAddr)
	fwd.Closed = true
}
