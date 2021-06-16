package tunnel

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
	"github.com/supergiant-hq/xnet/udp"
)

type TCPTunneler struct {
	forwards *sync.Map
	mutex    sync.Mutex
	log      *logrus.Logger
}

func NewTCPTunneler(log *logrus.Logger) *TCPTunneler {
	return &TCPTunneler{
		log:      log,
		forwards: new(sync.Map),
	}
}

func (tun *TCPTunneler) ForwardFrom(c *p2pc.Connection, fromAddr, toAddr string) (err error) {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	if forward, ok := tun.forwards.LoadAndDelete(fromAddr); ok {
		forward.(*Forward).Close("New forward request")
	}

	listenAddr, err := net.ResolveTCPAddr("tcp", fromAddr)
	if err != nil {
		return
	}

	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return
	}

	forward := &Forward{
		FromAddr:    fromAddr,
		ToAddr:      toAddr,
		Conn:        c,
		Forwards:    tun.forwards,
		TCPListener: listener,
	}
	tun.forwards.Store(fromAddr, forward)
	tun.log.Infof("-> [TCP] Forwarding connections from (%s)", fromAddr)

	go func() {
		defer func() {
			forward.Close("Connection closed")
		}()

		for {
			if c.Closed {
				break
			}

			conn, err := listener.Accept()
			if err != nil {
				tun.log.Errorln("[TCP] Error accepting connection on: ", fromAddr, err.Error())
				continue
			}
			tun.log.Infof("-> [TCP] Forwarding connection from (%s) to (%s)", fromAddr, toAddr)

			stream, err := c.OpenStream(map[string]string{
				"type":    TunTypeTCP,
				"address": toAddr,
			})
			if err != nil {
				tun.log.Errorln(err.Error())
				conn.Close()
				continue
			}

			go func() {
				defer conn.Close()
				_, err := io.Copy(conn, stream.Stream())
				if err != nil && err != io.EOF {
					tun.log.Errorln(err.Error())
				}
			}()

			go func() {
				defer stream.Close()
				_, err := io.Copy(stream.Stream(), conn)
				if err != nil && err != io.EOF {
					tun.log.Errorln(err.Error())
				}
			}()
		}
	}()

	return
}

func (tun *TCPTunneler) ForwardTo(stream *udp.Stream, address string) (err error) {
	raddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return
	}

	forwardConn, err := net.Dial("tcp", raddr.String())
	if err != nil {
		tun.log.Warnf("[TCP] Cannot open connection to %s: %s", address, err.Error())
		return
	}
	tun.log.Infof("<- [TCP] Forwarding connection to (%s)", address)

	go func() {
		defer stream.Close()
		_, err := io.Copy(stream.Stream(), forwardConn)
		if err != nil && err != io.EOF {
			tun.log.Errorln(err.Error())
		}
	}()

	go func() {
		defer forwardConn.Close()
		_, err := io.Copy(forwardConn, stream.Stream())
		if err != nil && err != io.EOF {
			tun.log.Errorln(err.Error())
		}
	}()

	return
}

func (tun *TCPTunneler) CloseConns() {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	tun.forwards.Range(func(key, value interface{}) bool {
		value.(*Forward).Close("Closing all connections")
		tun.forwards.Delete(key)
		return true
	})
}
