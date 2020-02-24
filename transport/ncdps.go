/*
	Author: Bo Maryniuk
	Node Controller daemon
*/
package ncdtransport

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"strings"
)

type NatsURL struct {
	Scheme string
	Fqdn   string
	Port   int
}

type NcdPubSub struct {
	urls []*NatsURL
	ncp  *nats.Conn
	ncs  *nats.Conn
}

func NewNcdPubSub() *NcdPubSub {
	ncd := new(NcdPubSub)
	ncd.urls = make([]*NatsURL, 0)
	return ncd
}

// AddNatsServerURL adds NATS server URL to the cluster of servers to connect
func (ncd *NcdPubSub) AddNatsServerURL(host string, port int) *NcdPubSub {
	ncd.urls = append(ncd.urls, &NatsURL{Scheme: "nats", Fqdn: host, Port: port})
	return ncd
}

// IsConnected currently only indicates if the connection is initialised
func (ncd *NcdPubSub) IsConnected() bool {
	return ncd.ncp != nil && ncd.ncs != nil
}

// Format cluster URLs
func (ncd *NcdPubSub) getClusterURLs() string {
	buff := make([]string, 0)
	for _, nurl := range ncd.urls {
		buff = append(buff, fmt.Sprintf("%s://%s:%d", nurl.Scheme, nurl.Fqdn, nurl.Port))
	}
	return strings.Join(buff, ", ")
}

// Connect to the cluster
func (ncd *NcdPubSub) connect() {
	var err error
	log.Printf("Connecting to %s...", ncd.getClusterURLs())
	if !ncd.IsConnected() {
		ncd.ncp, err = nats.Connect(ncd.getClusterURLs())
		log.Print("Connected publisher")
		if err != nil {
			log.Fatal(err)
		}
		ncd.ncs, err = nats.Connect(ncd.getClusterURLs())
		log.Print("Connected subscriber")
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Disconnect from the cluster
func (ncd *NcdPubSub) Disconnect() {
	if ncd.IsConnected() {
		log.Print("Begin disconnect")
		for _, nc := range [2]*nats.Conn{ncd.ncp, ncd.ncs} {
			if err := nc.Drain(); err != nil {
				log.Println(err.Error())
			}
			nc.Close()
		}
		ncd.ncp = nil
		ncd.ncs = nil
		log.Print("Disconected")
	}
}

func (ncd *NcdPubSub) GetPublisher() *nats.Conn {
	return ncd.ncp
}
func (ncd *NcdPubSub) GetSubscriber() *nats.Conn {
	return ncd.ncs
}

// Start starts the Node Controller
func (ncd *NcdPubSub) Start() {
	log.Print("Starting ncd event listener...")
	ncd.connect()
}
