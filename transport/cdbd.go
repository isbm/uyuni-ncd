package ncdtransport

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
)

type CdtCallback func(data map[string]interface{})

type CdtTransport struct {
	topic     string
	callbacks []CdtCallback
}

// Constructor
func NewCdtTransport(topic string) *CdtTransport {
	cdt := new(CdtTransport)
	cdt.topic = topic
	cdt.callbacks = make([]CdtCallback, 0)
	return cdt
}

// OnReceive is triggered by MQ when the message arrives
func (cdt *CdtTransport) OnReceive(body *nats.Msg) {
	var data interface{}
	if err := json.Unmarshal(body.Data, &data); err != nil {
		log.Println("ERROR: wrong message body -", err.Error())
	} else {
		for _, callback := range cdt.callbacks {
			callback(data.(map[string]interface{}))
		}
	}
}

// Channel returns subscribed topic
func (cdt *CdtTransport) Topic() string {
	return cdt.topic
}

// AddCallback adds an arbitrary callback, implementing transport.CdtCallback type.
func (cdt *CdtTransport) AddCallback(callback CdtCallback) {
	cdt.callbacks = append(cdt.callbacks, callback)
}
