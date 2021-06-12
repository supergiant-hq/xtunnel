package config

type CLIConfig struct {
	Mode             string
	BrokerListenAddr string
	RelayListenAddr  string
	BrokerAddr       string
	RelayAddr        string
	Token            string
	PeerID           string
	PeerMode         string
	TunType          string
	TunRev           bool
	TunFrom          string
	TunTo            string
	ConfigFile       string
	Debug            bool
}
