package tunnel

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
	"github.com/supergiant-hq/xnet/udp"
)

type UDPConn struct {
	udpConn  *net.UDPConn
	clientIp *net.UDPAddr
	stream   *udp.Stream
	log      *logrus.Logger
}

func (c *UDPConn) relayIncomingPackets() {
	defer func() {
		recover()
		c.stream.Close()
	}()

	packet := make([]byte, 2048)
	for {
		if c.stream.Closed {
			return
		}

		plen, err := c.stream.Stream().Read(packet)
		if err != nil {
			c.log.Errorln(err)
			break
		}
		_, err = c.udpConn.WriteToUDP(packet[:plen], c.clientIp)
		if err != nil {
			c.log.Errorf("[UDP] Read from remote error: %+v", err)
		}
	}
}

type UDPTunnel struct {
	conns map[string]*UDPConn
	mutex sync.RWMutex
	log   *logrus.Logger
}

func NewUDPTunnel(log *logrus.Logger) *UDPTunnel {
	return &UDPTunnel{
		conns: make(map[string]*UDPConn),
		log:   log,
	}
}

func (tun *UDPTunnel) ForwardFrom(c *p2pc.Connection, fromAddr, toAddr string) (listener *net.UDPConn, err error) {
	listenAddr, err := net.ResolveUDPAddr("udp", fromAddr)
	if err != nil {
		return
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", toAddr)
	if err != nil {
		return
	}

	listener, err = net.ListenUDP("udp", listenAddr)
	if err != nil {
		return
	}
	tun.log.Infof("-> [UDP] Forwarding packets from (%s)", listenAddr.String())

	go func() {
		defer func() {
			recover()
			listener.Close()
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

func (tun *UDPTunnel) ForwardTo(stream *udp.Stream, toAddr string) (err error) {
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
