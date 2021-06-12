package client

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/supergiant-hq/tunnel/config"
	tunModel "github.com/supergiant-hq/tunnel/model"
	"github.com/supergiant-hq/tunnel/tunnel"
	"github.com/supergiant-hq/xnet/p2p"
	brokerc "github.com/supergiant-hq/xnet/p2p/broker/client"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
	"github.com/supergiant-hq/xnet/udp"
	udpc "github.com/supergiant-hq/xnet/udp/client"
)

type client struct {
	cfg       config.CLIConfig
	tcpTunnel *tunnel.TCPTunnel
	udpTunnel *tunnel.UDPTunnel
	log       *logrus.Logger
}

func LaunchClient(log *logrus.Logger, cfg config.CLIConfig) (c *brokerc.Client, err error) {
	log.Infoln("Launching Client...")

	client := client{
		cfg:       cfg,
		tcpTunnel: tunnel.NewTCPTunnel(log),
		udpTunnel: tunnel.NewUDPTunnel(log),
		log:       log,
	}

	serverAddr, err := net.ResolveUDPAddr("udp", cfg.BrokerAddr)
	if err != nil {
		return
	}

	var relayUdpAddr *net.UDPAddr
	if len(cfg.RelayAddr) > 0 {
		relayUdpAddr, err = net.ResolveUDPAddr("udp", cfg.RelayAddr)
		if err != nil {
			return
		}
	}

	if c, err = brokerc.New(
		brokerc.Config{
			Debug: cfg.Debug,
			UdpcConfig: udpc.Config{
				Tag:         "CLI",
				ServerAddr:  serverAddr,
				Token:       cfg.Token,
				Unmarshaler: tunModel.Unmarshal,
			},
			P2PConfig: p2pc.Config{
				RelayAddr: relayUdpAddr,
			},
		},
		client.onNewStream,
	); err != nil {
		return
	}

	c.SetMessageStreamHandler(client.onNewMessageStream)

	if err = c.Connect(); err != nil {
		return
	}

	if len(cfg.PeerID) > 0 {
		client.connectToPeer(c, cfg)
	}

	return
}

func (c *client) connectToPeer(bc *brokerc.Client, cfg config.CLIConfig) {
	for {
		conn, err := bc.ConnectPeerById(cfg.PeerID, p2p.ConnectionMode(cfg.PeerMode))
		if err != nil {
			c.log.Errorln(err.Error())
			continue
		}

		if cfg.TunRev {
			if err = tunnel.Reverse(c.log, conn, cfg); err != nil {
				c.log.Errorln(err.Error())
				conn.Close(err.Error())
				continue
			} else {
				<-conn.Exit
			}
		} else {
			switch cfg.TunType {
			case tunnel.TunTypeTCP:
				if listener, err := c.tcpTunnel.ForwardFrom(conn, cfg.TunFrom, cfg.TunTo); err != nil {
					c.log.Errorln(err.Error())
					conn.Close(err.Error())
					continue
				} else {
					<-conn.Exit
					listener.Close()
				}
			case tunnel.TunTypeUDP:
				if listener, err := c.udpTunnel.ForwardFrom(conn, cfg.TunFrom, cfg.TunTo); err != nil {
					c.log.Errorln(err.Error())
					conn.Close(err.Error())
					continue
				} else {
					<-conn.Exit
					listener.Close()
				}
			default:
				c.log.Errorln("Invalid TunType")
				conn.Close("Invalid TunType")
				continue
			}
		}
	}

}

func (c *client) onNewStream(client udp.Client, stream *udp.Stream) {
	switch stream.Data["type"] {
	case tunnel.TunTypeTCP:
		if err := c.tcpTunnel.ForwardTo(stream, stream.Data["address"]); err != nil {
			c.log.Errorln(err.Error())
			stream.Close()
		}
	case tunnel.TunTypeUDP:
		if err := c.udpTunnel.ForwardTo(stream, stream.Data["address"]); err != nil {
			c.log.Errorln(err.Error())
			stream.Close()
		}
	default:
		c.log.Errorln("Invalid TunType")
		stream.Close()
	}
}

func (c *client) onNewMessageStream(ms *p2pc.MessageStream) {
	if err := ms.Listen(c.messageHandler); err != nil {
		ms.Close()
		return
	}
}
