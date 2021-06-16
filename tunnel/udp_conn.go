package tunnel

import (
	"net"

	"github.com/sirupsen/logrus"
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
