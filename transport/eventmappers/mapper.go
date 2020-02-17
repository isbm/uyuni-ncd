package eventmappers

import (
	"github.com/isbm/uyuni-ncd/transport"
)

type Mapper interface {
	Label() string
	TopicRoot() string
	OnMQReceive(m *ncdtransport.MqMessage)
	OnIntReceive(m *ncdtransport.InternalEventMessage) *ncdtransport.MqMessage
}
