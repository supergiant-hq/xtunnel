package config

type CLIConfig struct {
	Mode             string
	BrokerListenAddr string
	RelayListenAddr  string
	BrokerAddr       string
	RelayAddr        string
	Token            string
	TunPeer          string
	TunPeerMode      string
	TunType          string
	TunRev           bool
	TunFrom          string
	TunTo            string
	Debug            bool
}
