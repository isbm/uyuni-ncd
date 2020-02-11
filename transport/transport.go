package ncdtransport

import (
	"github.com/nats-io/nats.go"
)

// Subscriber interface
type Subscriber interface {
	OnReceive(body *nats.Msg)
	Topic() string
}

// Publisher interface
type Publisher interface {
	SetPublisher(nc *nats.Conn)
	Topic() string
}
