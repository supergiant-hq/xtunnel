package tunnel

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/supergiant-hq/tunnel/config"
	model "github.com/supergiant-hq/tunnel/model"

	"github.com/supergiant-hq/xnet/network"
	p2pc "github.com/supergiant-hq/xnet/p2p/client"
)

func Reverse(log *logrus.Logger, conn *p2pc.Connection, cfg config.Tunnel) (err error) {
	log.Infoln("Requesting reverse tunnel...")

	ms, err := conn.OpenMessageStream()
	if err != nil {
		return
	}
	defer ms.Close()

	msgData := &model.TunnelOpen{
		Type:        cfg.Type,
		FromAddress: cfg.From,
		ToAddress:   cfg.To,
	}
	msg := network.NewMessageWithAck(model.MessageTypeTunnelOpen, msgData, network.RequestTimeout)
	rmsg, err := ms.SendAndRead(msg)
	if err != nil {
		return
	}
	rdata := rmsg.Body.(*model.TunnelOpenResponse)
	if !rdata.Status {
		err = fmt.Errorf(rdata.Message)
		return
	}

	log.Infof("Reverse tunnel established from(%v) to(%v) type(%v)", msgData.FromAddress, msgData.ToAddress, msgData.Type)

	return
}
