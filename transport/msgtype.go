package ncdtransport

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"strings"
)

/*
MqMessage can drive various topics.

"Action" is an arbitrary convention-based label. For databases
it is "insert", "delete" or "update", which reflects the trigger.
Other mesages could have other topics.

"Topic" is a mapping path of message interpreter and the topic itself.
For databases it is "/db/<topic>". For example, to update or add
or remove a channel, the topic is "/db/channel". Other topic might
be management of a node, so it can be e.g. "/cfg" which would mean
that the "Payload" is a configuration management nanostate and should
be passed down to the nanostate interpreter for further processing.
*/
type MqMessage struct {
	Id      string
	Action  string
	Topic   string
	Payload interface{}
}

func NewMqMessage() *MqMessage {
	msg := new(MqMessage)
	msg.Id = uuid.New().String()

	return msg
}

// Load self content from given bytes
func (bm *MqMessage) FromBytes(data []byte) *MqMessage {
	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		log.Panicln("Error loading incoming JSON:", err.Error())
	}
	for section, obj := range content {
		switch section {
		case "Topic":
			bm.Topic = obj.(string)
		case "Payload":
			bm.Payload = obj
		case "Id":
			bm.Id = obj.(string)
		case "Action":
			bm.Action = obj.(string)
		default:
			log.Panicln("Unknown type section:", section)
		}
	}
	return bm
}

// Serialise this object to bytes
func (bm *MqMessage) ToBytes() []byte {
	data, err := json.Marshal(&bm)
	if err != nil {
		panic(err)
	}
	return data
}

// Serialise this object to JSON string
func (bm *MqMessage) ToJSON() string {
	return string(bm.ToBytes())
}

type InternalEventMessage struct {
	Payload map[string]interface{}
	Topic   string
	Action  string
}

func NewInternalEventMessage(data map[string]interface{}) *InternalEventMessage {
	dem := new(InternalEventMessage)
	dem.Topic = data["table"].(string)
	dem.Action = strings.ToLower(data["action"].(string))
	dem.Payload = data["data"].(map[string]interface{})

	return dem
}

func (iem *InternalEventMessage) FromData(data map[string]interface{}) *InternalEventMessage {
	for section, obj := range data {
		switch section {
		case "Topic":
			iem.Topic = obj.(string)
		case "Payload":
			iem.Payload = obj.(map[string]interface{})
		case "Action":
			iem.Action = obj.(string)
		default:
			log.Panicln("Unknown type section:", section)
		}
	}
	return iem
}

// Load self content from given bytes
func (iem *InternalEventMessage) FromBytes(data []byte) *InternalEventMessage {
	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		log.Panicln("Error loading incoming JSON:", err.Error())
	}
	return iem.FromData(content)
}
