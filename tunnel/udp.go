package tunnel

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
	"github.com/supergiant-hq/xnet/udp"
)

type UDPTunneler struct {
	conns    map[string]*UDPConn
	forwards *sync.Map
	mutex    sync.RWMutex
	log      *logrus.Logger
}

func NewUDPTunneler(log *logrus.Logger) *UDPTunneler {
	return &UDPTunneler{
		conns:    make(map[string]*UDPConn),
		forwards: new(sync.Map),
		log:      log,
	}
}

func (tun *UDPTunneler) ForwardFrom(c *p2pc.Connection, fromAddr, toAddr string) (err error) {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	if forward, ok := tun.forwards.LoadAndDelete(fromAddr); ok {
		forward.(*Forward).Close("New forward request")
	}

	listenAddr, err := net.ResolveUDPAddr("udp", fromAddr)
	if err != nil {
		return
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", toAddr)
	if err != nil {
		return
	}

	listener, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return
	}

	forward := &Forward{
		FromAddr:    fromAddr,
		ToAddr:      toAddr,
		Conn:        c,
		Forwards:    tun.forwards,
		UDPListener: listener,
	}
	tun.forwards.Store(fromAddr, forward)
	tun.log.Infof("-> [UDP] Forwarding packets from (%s)", listenAddr.String())

	go func() {
		defer func() {
			forward.Close("Connection closed")
		}()

		for {
			if c.Closed {
				break
			}

			packet := make([]byte, 2048)
			plen, paddr, err := listener.ReadFromUDP(packet)
			if err != nil {
				tun.log.Errorln(err)
				break
			}
			tun.log.Debugf("[UDP] Sending to remote from addr: %+v (%+v)", paddr, err)

			tun.mutex.RLock()
			conn, ok := tun.conns[paddr.String()]
			tun.mutex.RUnlock()
			if !ok {
				stream, err := c.OpenStream(map[string]string{
					"type":    TunTypeUDP,
					"address": toAddr,
				})
				if err != nil {
					tun.log.Warnf("-> [UDP] Error creating stream from (%s) to (%s)", listenAddr.String(), remoteAddr.String())
					continue
				}

				tun.mutex.Lock()
				conn = &UDPConn{
					udpConn:  listener,
					clientIp: paddr,
					stream:   stream,
					log:      tun.log,
				}
				tun.conns[conn.clientIp.String()] = conn
				go conn.relayIncomingPackets()
				tun.mutex.Unlock()

				tun.log.Debugf("-> [UDP] Forwarding stream from (%s) to (%s)", listenAddr.String(), remoteAddr.String())
			}

			_, err = conn.stream.Stream().Write(packet[:plen])
			if err != nil {
				tun.log.Errorf("[UDP] Sending to remote error: %+v", err)
				tun.mutex.Lock()
				delete(tun.conns, conn.clientIp.String())
				tun.mutex.Unlock()
			}
		}
	}()

	return
}

func (tun *UDPTunneler) ForwardTo(stream *udp.Stream, toAddr string) (err error) {
	remoteAddr, err := net.ResolveUDPAddr("udp", toAddr)
	if err != nil {
		return
	}

	forwardConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		tun.log.Errorf("[UDP] Cannot dial connection to %s: %s", remoteAddr.String(), err.Error())
		return
	}
	tun.log.Infof("<- [UDP] Forwarding packets to (%s)", remoteAddr.String())

	go func() {
		defer forwardConn.Close()
		_, err := io.Copy(forwardConn, stream.Stream())
		if err != io.EOF {
			tun.log.Errorln(err.Error())
		}
	}()

	go func() {
		defer stream.Close()
		_, err := io.Copy(stream.Stream(), forwardConn)
		if err != io.EOF {
			tun.log.Errorln(err.Error())
		}
	}()

	<-stream.Exit

	return
}

func (tun *UDPTunneler) CloseConns() {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	tun.forwards.Range(func(key, value interface{}) bool {
		value.(*Forward).Close("Closing all connections")
		tun.forwards.Delete(key)
		return true
	})
}
