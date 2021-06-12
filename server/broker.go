package server

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/supergiant-hq/tunnel/config"
	tunModel "github.com/supergiant-hq/tunnel/model"
	"github.com/supergiant-hq/xnet/model"
	"github.com/supergiant-hq/xnet/p2p"
	brokers "github.com/supergiant-hq/xnet/p2p/broker/server"
	udps "github.com/supergiant-hq/xnet/udp/server"
)

func LaunchBroker(log *logrus.Logger, cfg config.CLIConfig) (s *brokers.Server, err error) {
	log.Infoln("Launching Broker...")

	fileCfg, err := config.ParseFile(cfg.ConfigFile)
	if err != nil {
		err = fmt.Errorf("error parsing config file: %v", err.Error())
		return
	}

	listenAddr, err := net.ResolveUDPAddr("udp", cfg.BrokerListenAddr)
	if err != nil {
		return
	}

	if s, err = brokers.New(
		brokers.Config{
			Debug: cfg.Debug,
			UdpsConfig: udps.Config{
				Tag:         "CLI",
				Addr:        listenAddr,
				Unmarshaler: tunModel.Unmarshal,
			},
		},
		GetClientValidator(fileCfg),
	); err != nil {
		return
	}

	if err = s.Listen(); err != nil {
		return
	}

	return
}

func GetClientValidator(cfg *config.FileConfig) udps.ClientValidateHandler {
	return func(u *net.UDPAddr, cvd *model.ClientValidateData) (cd *model.ClientData, err error) {
		client, err := cfg.GetClientFromToken(cvd.Token)
		if err != nil {
			return
		}

		tags := map[string]string{
			udps.KEY_CLIENT_ID: client.Id,
		}
		if client.Relay {
			tags[p2p.TAG_RELAY] = "true"
		}

		cd = &model.ClientData{
			Id:      client.Id,
			Address: u.String(),
			Tags:    tags,
			Data:    cvd.Data,
			Ctx: &model.ClientData_BrokerCtx{
				BrokerCtx: &model.BrokerClientContext{},
			},
		}

		return
	}
}
