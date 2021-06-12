package server

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/supergiant-hq/tunnel/config"
	relay "github.com/supergiant-hq/xnet/p2p/relay"
)

func LaunchRelay(log *logrus.Logger, cfg config.CLIConfig) (s *relay.Server, err error) {
	log.Infoln("Launching Relay...")

	listenAddr, err := net.ResolveUDPAddr("udp", cfg.RelayListenAddr)
	if err != nil {
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp", cfg.BrokerAddr)
	if err != nil {
		return
	}

	if s, err = relay.New(
		relay.Config{
			Debug:       cfg.Debug,
			Addr:        listenAddr,
			BrokerAddr:  serverAddr,
			BrokerToken: cfg.Token,
		},
	); err != nil {
		return
	}

	if err = s.Listen(); err != nil {
		return
	}

	return
}
