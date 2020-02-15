package eventmappers

import (
	"github.com/isbm/uyuni-ncd/transport"
)

type Mapper interface {
	Label() string
	TopicRoot() string
	OnReceive(m ncdtransport.MqMessage)
}
