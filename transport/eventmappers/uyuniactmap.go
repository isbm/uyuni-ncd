package eventmappers

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/isbm/uyuni-ncd/transport"
)

type ActionFunc func(m *ncdtransport.MqMessage) error

type UyuniActionsMap struct {
	mapper *UyuniEventMapper
	fmap   map[string]ActionFunc
}

func NewUyuniActionsMap(uem *UyuniEventMapper) *UyuniActionsMap {
	uam := new(UyuniActionsMap)
	uam.mapper = uem
	uam.fmap = map[string]ActionFunc{
		"/uyuni/rhnchannel": uam.onRhnChannel,
	}
	return uam
}

func (uam *UyuniActionsMap) OnTopic(m *ncdtransport.MqMessage) error {
	call, ex := uam.fmap[m.Topic]
	if !ex {
		return fmt.Errorf("No actionable topic '%s' has been found", m.Topic)
	}

	return call(m)
}

func (uam *UyuniActionsMap) onRhnChannel(m *ncdtransport.MqMessage) error {
	switch m.Topic {
	case "/uyuni/rhnchannel":
		/*
			Every entity should be always created and then updated.
			However creation step should be omitted, if the entity wasn't found on boot index.
		*/
		switch m.Action {
		case "update":
			fmt.Println("Action", m.Action)
			spew.Dump(m.Payload)
			args := make([]interface{}, 0)

			for _, arg := range []string{"label", "name", "summary", "arch_label",
				"parent_channel_label", "checksum_label", "gpgkey", "gpg_check"} {
				if arg == "gpgkey" {
					args = append(args, map[string]interface{}{
						"url":         m.Payload.(map[string]interface{})["gpg_key_url"],
						"id":          m.Payload.(map[string]interface{})["gpg_key_id"],
						"fingerprint": m.Payload.(map[string]interface{})["gpg_key_fp"],
					})
				} else {
					args = append(args, m.Payload.(map[string]interface{})[arg])
				}
			}
			uam.mapper.scall("channel.software.create", args...)
		}
	}
	return nil
}
