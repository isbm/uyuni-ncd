/*
Internal events reactor mapper. It is designed to fetch required data via XML-RPC API
from the Uyuni Server on certain events and prepare a payload for the MQ Message.
*/

package eventmappers

import (
	"fmt"
	"github.com/isbm/uyuni-ncd/transport"
	"log"
)

type MapFunc func(action string, data map[string]interface{}) interface{}

type UyuniIntMap struct {
	mapper *UyuniEventMapper
	fmap   map[string]MapFunc
}

func NewUyuniIntMap(uem *UyuniEventMapper) *UyuniIntMap {
	uim := new(UyuniIntMap)
	uim.mapper = uem
	uim.fmap = map[string]MapFunc{
		"rhnchannel": uim.onRhnChannel,
	}
	return uim
}

func (uim *UyuniIntMap) OnTopic(m *ncdtransport.InternalEventMessage) (interface{}, error) {
	call, ex := uim.fmap[m.Topic]
	if !ex {
		return nil, fmt.Errorf("No topic '%s' has been found", m.Topic)
	}

	return call(m.Action, m.Payload), nil
}

///////////////// Mappers

// Action for "rhnchannel" table
func (uim *UyuniIntMap) onRhnChannel(action string, data map[string]interface{}) interface{} {
	var out interface{}
	switch action {
	case "insert":
		// Explicitly ignore. It is always an update afterwards.
	case "update":
		out = uim.mapper.scall("channel.software.getDetails", data["label"].(string))
	case "delete":
		out = data["label"]
	default:
		log.Println("No destination defined on action", action)
	}
	return out
}
