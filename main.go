package main

import (
	"flag"

	"github.com/supergiant-hq/tunnel/client"
	"github.com/supergiant-hq/tunnel/config"
	"github.com/supergiant-hq/tunnel/server"
	"github.com/supergiant-hq/xnet/p2p"
	"github.com/supergiant-hq/xnet/util"

	"github.com/sirupsen/logrus"
)

var (
	mode             = flag.String("mode", "broker", "broker, relay or client")
	brokerListenAddr = flag.String("brokerListen", ":10000", "Broker Listen Address")
	relayListenAddr  = flag.String("relayListen", ":15000", "Relay Listen Address")
	brokerAddr       = flag.String("broker", ":10000", "Broker Address")
	relayAddr        = flag.String("relay", "", "Relay Address")
	token            = flag.String("token", "", "Auth Token")
	peerId           = flag.String("peerid", "", "client-1")
	peerMode         = flag.String("peermode", string(p2p.ConnectionModeP2P), "p2p or relay")
	tunType          = flag.String("tuntype", "tcp", "tcp or udp")
	tunRev           = flag.Bool("tunrev", false, "Reverse Tunnel")
	tunFrom          = flag.String("tunfrom", "", "Tunnel from ip:port")
	tunTo            = flag.String("tunto", "", "Tunnel to ip:port")
	configFile       = flag.String("config", "./config.yaml", "Config File")
	debug            = flag.Bool("d", false, "Debug")
)

func main() {
	flag.Parse()

	logLevel := logrus.InfoLevel
	if *debug {
		logLevel = logrus.DebugLevel
	}
	log := util.NewLogger(logLevel)

	cliCfg := config.CLIConfig{
		Mode:             *mode,
		BrokerListenAddr: *brokerListenAddr,
		RelayListenAddr:  *relayListenAddr,
		BrokerAddr:       *brokerAddr,
		RelayAddr:        *relayAddr,
		Token:            *token,
		PeerID:           *peerId,
		PeerMode:         *peerMode,
		TunType:          *tunType,
		TunRev:           *tunRev,
		TunFrom:          *tunFrom,
		TunTo:            *tunTo,
		ConfigFile:       *configFile,
		Debug:            *debug,
	}

	switch *mode {
	case "broker":
		s, err := server.LaunchBroker(log, cliCfg)
		if err != nil {
			log.Fatalln(err)
		}
		<-s.Exit

	case "relay":
		s, err := server.LaunchRelay(log, cliCfg)
		if err != nil {
			log.Fatalln(err)
		}
		<-s.Exit

	case "client":
		c, err := client.LaunchClient(log, cliCfg)
		if err != nil {
			log.Fatalln(err)
		}
		<-c.Exit

	default:
		log.Fatalln("Invalid Mode")
		return
	}
}
