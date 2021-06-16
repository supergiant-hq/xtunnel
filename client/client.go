package client

import (
	"net"
	"time"

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
	cfg         config.CLIConfig
	tcpTunneler *tunnel.TCPTunneler
	udpTunneler *tunnel.UDPTunneler
	log         *logrus.Logger
}

func LaunchClient(log *logrus.Logger, cfg config.CLIConfig, fileCfg config.FileConfig) (c *brokerc.Client, err error) {
	log.Infoln("Launching Client...")

	client := client{
		cfg:         cfg,
		tcpTunneler: tunnel.NewTCPTunneler(log),
		udpTunneler: tunnel.NewUDPTunneler(log),
		log:         log,
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

	c.SetDisconnectedHandler(func() {
		client.tcpTunneler.CloseConns()
		client.udpTunneler.CloseConns()
	})
	c.SetMessageStreamHandler(client.onNewMessageStream)

	if err = c.Connect(); err != nil {
		return
	}

	tunnels := fileCfg.Tunnels
	if len(tunnels) == 0 && len(cfg.TunPeer) > 0 {
		tunnels = append(tunnels, config.Tunnel{
			Peer:    cfg.TunPeer,
			Mode:    cfg.TunPeerMode,
			Type:    cfg.TunType,
			Reverse: cfg.TunRev,
			From:    cfg.TunFrom,
			To:      cfg.TunTo,
		})
	}

	if len(tunnels) > 0 {
		client.log.Infoln("Establishing tunnels...")
		for _, tunCfg := range tunnels {
			go client.connectToPeer(c, tunCfg)
		}
	}

	return
}

func (c *client) connectToPeer(bc *brokerc.Client, cfg config.Tunnel) {
	for {
		<-time.After(time.Second)

		conn, err := bc.ConnectPeerById(cfg.Peer, p2p.ConnectionMode(cfg.Mode))
		if err != nil {
			c.log.Errorln(err.Error())
			continue
		}

		if cfg.Reverse {
			if err = tunnel.Reverse(c.log, conn, cfg); err != nil {
				c.log.Errorln(err.Error())
				conn.Close(err.Error())
				continue
			} else {
				<-conn.Exit
			}
		} else {
			switch cfg.Type {
			case tunnel.TunTypeTCP:
				if err := c.tcpTunneler.ForwardFrom(conn, cfg.From, cfg.To); err != nil {
					c.log.Errorln(err.Error())
					conn.Close(err.Error())
					continue
				} else {
					<-conn.Exit
				}
			case tunnel.TunTypeUDP:
				if err := c.udpTunneler.ForwardFrom(conn, cfg.From, cfg.To); err != nil {
					c.log.Errorln(err.Error())
					conn.Close(err.Error())
					continue
				} else {
					<-conn.Exit
				}
			default:
				c.log.Errorln("Invalid TunType")
				conn.Close("Invalid TunType")
				return
			}
		}
	}
}

func (c *client) onNewStream(client udp.Client, stream *udp.Stream) {
	switch stream.Data["type"] {
	case tunnel.TunTypeTCP:
		if err := c.tcpTunneler.ForwardTo(stream, stream.Data["address"]); err != nil {
			c.log.Errorln(err.Error())
			stream.Close()
		}
	case tunnel.TunTypeUDP:
		if err := c.udpTunneler.ForwardTo(stream, stream.Data["address"]); err != nil {
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
