package client

import (
	"fmt"

	"github.com/supergiant-hq/tunnel/model"
	"github.com/supergiant-hq/tunnel/tunnel"
	"github.com/supergiant-hq/xnet/network"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
)

func (c *client) messageHandler(ms *p2pc.MessageStream, msg *network.Message) {
	switch msg.Ctx.Type {
	case model.MessageTypeTunnelOpen:
		c.tunnelOpenHandler(ms, msg)
	default:
		c.log.Errorf("MessageHandler for type (%v) not found", msg.Ctx.Type)
	}
}

func (c *client) tunnelOpenHandler(ms *p2pc.MessageStream, msg *network.Message) {
	var err error

	msgData := msg.Body.(*model.TunnelOpen)
	c.log.Infof("Open tunnel request: %+v", msgData)

	defer func() {
		resData := &model.TunnelOpenResponse{}

		if err != nil {
			resData.Status = false
			resData.Message = err.Error()
			c.log.Errorf("Error accepting tunnel request: %+v", err.Error())
		} else {
			resData.Status = true
			resData.Message = "Ok"
			c.log.Infof("Tunnel established from(%v) to(%v) type(%v)", msgData.FromAddress, msgData.ToAddress, msgData.Type)
		}

		mres, _ := msg.GenReply(model.MessageTypeTunnelOpenResponse, resData)
		if _, err := ms.Send(mres); err != nil {
			c.log.Errorf("Error replying to tunnel request: %+v", err)
		}
	}()

	switch msgData.Type {
	case tunnel.TunTypeTCP:
		err = c.tcpTunneler.ForwardFrom(ms.Conn, msgData.FromAddress, msgData.ToAddress)
	case tunnel.TunTypeUDP:
		err = c.udpTunneler.ForwardFrom(ms.Conn, msgData.FromAddress, msgData.ToAddress)
	default:
		err = fmt.Errorf("invalid tunnel mode")
	}
}
