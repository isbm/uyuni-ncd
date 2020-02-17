// The ncd itself

package ncd

import (
	"fmt"
	"github.com/isbm/uyuni-ncd/transport"
	"github.com/isbm/uyuni-ncd/transport/eventmappers"
	"github.com/nats-io/nats.go"
	"log"
	"strings"
)

const (
	CHANNEL_NODES    = "nodes"
	CHANNEL_DIRECTOR = "director"
)

type NcdConf struct {
	Running bool
	Leader  bool
}

type Ncd struct {
	rtconf    *NcdConf
	transport *ncdtransport.NcdPubSub
	dbl       *ncdtransport.PgEventListener
	reflector *ncdtransport.MsgIdBuff
	_mappers  []*eventmappers.Mapper
}

func NewNcd() *Ncd {
	n := new(Ncd)
	n.rtconf = &NcdConf{}
	n.transport = ncdtransport.NewNcdPubSub()
	n.dbl = ncdtransport.NewPgEventListener()
	n.reflector = ncdtransport.NewMsgIdBuff()
	n._mappers = make([]*eventmappers.Mapper, 0)

	return n
}

// AddMapper adds a mapper to the ncd
func (n *Ncd) AddMapper(mapper eventmappers.Mapper) *Ncd {
	n._mappers = append(n._mappers, &mapper)
	return n
}

// GetMapper by a topic
func (n *Ncd) GetMapper(topic string) (*eventmappers.Mapper, error) {
	for _, mobj := range n._mappers {
		mapper := *mobj
		if strings.HasPrefix(topic, mapper.TopicRoot()) {
			return mobj, nil
		}
	}
	return nil, fmt.Errorf("No mapper found for '%s' topic", topic)
}

// GetTransport returns NcdPubSub instance
func (n *Ncd) GetTransport() *ncdtransport.NcdPubSub {
	return n.transport
}

// GetDBListener return PgEventListener instance
func (n *Ncd) GetDBListener() *ncdtransport.PgEventListener {
	return n.dbl
}

// IsRunning returns true, if the Ncd is already running
func (n *Ncd) IsRunning() bool {
	return n.rtconf.Running
}

// IsLeader returns true, if the current node is a leader node
func (n *Ncd) IsLeader() bool {
	return n.rtconf.Leader
}

// SetLeader sets the current node into a leader mode
func (n *Ncd) SetLeader(leader bool) *Ncd {
	n.rtconf.Leader = leader
	return n
}

/////// Internal

// Handles CHANNEL_NODES inbox
func (n *Ncd) nodesHandler(m *nats.Msg) {
	log.Println("NH: received", len(m.Data), "bytes")
	msg := ncdtransport.NewMqMessage().FromBytes(m.Data)
	if n.reflector.Channel(CHANNEL_NODES).Discard(msg.Id) {
		mapper, err := n.GetMapper(msg.Topic)
		if err != nil {
			panic(err)
		}
		(*(mapper)).OnMQReceive(msg)
	}
}

// Handles CHANNELDIRECTOR inbox
func (n *Ncd) controllerHandler(m *nats.Msg) {
	fmt.Println("> from controller:", string(m.Data))
}

// XXX: Temporary handler for Uyuni Server database only. This should be moved to a plugin system.
// Handles DB external events
func (n *Ncd) externalHandler(m interface{}) {
	// Get Uyuni handler to deal with the database messages
	mpref, err := n.GetMapper("/uyuni")
	if err != nil {
		panic(err)
	}
	log.Println("EH: get mapper")

	switch (*(mpref)).Label() {
	case "UyuniEventMapper":
		uyuni := (*(mpref)).(*eventmappers.UyuniEventMapper)
		msg := uyuni.OnIntReceive(ncdtransport.NewInternalEventMessage(m.(map[string]interface{})))

		// send only if the current node is a leader and topic is supported
		if n.IsLeader() && msg.Topic != "" {
			err := n.GetTransport().GetPublisher().Publish(CHANNEL_NODES, msg.ToBytes())
			if err != nil {
				log.Panicln("Publishing error:", err)
			}
			n.reflector.Channel(CHANNEL_NODES).Push(msg.Id)
			log.Println("EH: Published to", CHANNEL_NODES)
		}
	}
}

// Internal, actual start.
func (n *Ncd) _start() {
	if n.IsRunning() {
		return
	}

	// Setup MQ
	n.GetTransport().Start()
	var s1 *nats.Subscription
	var s2 *nats.Subscription
	var err error
	s1, err = n.GetTransport().GetSubscriber().Subscribe(CHANNEL_NODES, n.nodesHandler)
	if err != nil {
		log.Panicln("Cannot subscribe to", CHANNEL_NODES, err.Error())
	}
	s2, err = n.GetTransport().GetSubscriber().Subscribe(CHANNEL_DIRECTOR, n.controllerHandler)
	if err != nil {
		log.Panicln("Cannot subscribe to", CHANNEL_DIRECTOR, err.Error())
	}

	fmt.Println(s1, s2)

	// Setup Db listener and start it in background
	// Dynamic design ideas:
	//   1. Implement as a plugin
	//   2. GetPlugins() -> []Plugin
	//   3. For each apply map of callbacks, or one common that distinguishes the desinations etc
	n.GetDBListener().AddCallback(n.externalHandler).Start()

	n.rtconf.Running = true
}

// Run ncd in background
func (n *Ncd) RunProcess() {
	go n._start()
}

// Run ncd in foreground
func (n *Ncd) Run() {
	n._start()
}

// Stop ncd
func (n *Ncd) Stop() {
	if err := n.GetTransport().GetSubscriber().Drain(); err != nil {
		panic("Drain error: " + err.Error())
	}
	n.rtconf.Running = false
}
