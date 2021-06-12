package model

import (
	"fmt"

	"github.com/supergiant-hq/xnet/network"
	"google.golang.org/protobuf/proto"
)

const (
	MessageTypeTunnelOpen         = network.MessageType("tunnel-open")
	MessageTypeTunnelOpenResponse = network.MessageType("tunnel-open-response")
)

// Unmarshal data based on the MessageType
func Unmarshal(mtype network.MessageType) (body proto.Message, err error) {
	switch mtype {
	case MessageTypeTunnelOpen:
		body = &TunnelOpen{}
	case MessageTypeTunnelOpenResponse:
		body = &TunnelOpenResponse{}
	default:
		err = fmt.Errorf("unmarshal proto model not found for type(%v)", mtype)
	}
	return
}
