package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Client struct {
	Id    string
	Token string
	Relay bool
}

type FileConfig struct {
	Clients []Client
}

func ParseFile(configFile string) (cfg *FileConfig, err error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return
	}

	cfg = &FileConfig{}
	if err = yaml.Unmarshal(bytes, cfg); err != nil {
		return
	}

	return
}

func (cfg *FileConfig) GetClientFromToken(token string) (client *Client, err error) {
	for _, c := range cfg.Clients {
		if c.Token == token {
			client = &c
			break
		}
	}

	if client == nil {
		err = fmt.Errorf("client not found")
	}

	return
}
