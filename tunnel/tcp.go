package tunnel

import (
	"io"
	"net"

	"github.com/sirupsen/logrus"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
	"github.com/supergiant-hq/xnet/udp"
)

type TCPTunnel struct {
	log *logrus.Logger
}

func NewTCPTunnel(log *logrus.Logger) *TCPTunnel {
	return &TCPTunnel{
		log: log,
	}
}

func (tun *TCPTunnel) ForwardFrom(c *p2pc.Connection, fromAddress, toAddress string) (listener *net.TCPListener, err error) {
	listenAddr, err := net.ResolveTCPAddr("tcp", fromAddress)
	if err != nil {
		return
	}

	listener, err = net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return
	}
	tun.log.Infof("-> [TCP] Forwarding connections from (%s)", fromAddress)

	go func() {
		defer listener.Close()

		for {
			if c.Closed {
				break
			}

			conn, err := listener.Accept()
			if err != nil {
				tun.log.Errorln("Error accepting connection on: ", fromAddress, err.Error())
				continue
			}
			tun.log.Infof("-> [TCP] Forwarding connection from (%s) to (%s)", fromAddress, toAddress)

			stream, err := c.OpenStream(map[string]string{
				"type":    TunTypeTCP,
				"address": toAddress,
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

func (tun *TCPTunnel) ForwardTo(stream *udp.Stream, address string) (err error) {
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
