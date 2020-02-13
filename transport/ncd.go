/*
	Author: Bo Maryniuk
	Node Controller daemon
*/
package ncd

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

type Ncd struct {
	urls []*NatsURL
	ncp  *nats.Conn
	ncs  *nats.Conn
}

func NewNcd() *Ncd {
	ncd := new(Ncd)
	ncd.urls = make([]*NatsURL, 0)
	return ncd
}

// AddNatsServerURL adds NATS server URL to the cluster of servers to connect
func (ncd *Ncd) AddNatsServerURL(host string, port int) *Ncd {
	ncd.urls = append(ncd.urls, &NatsURL{Scheme: "nats", Fqdn: host, Port: port})
	return ncd
}

// IsConnected currently only indicates if the connection is initialised
func (ncd *Ncd) IsConnected() bool {
	return ncd.ncp != nil && ncd.ncs != nil
}

// Format cluster URLs
func (ncd *Ncd) getClusterURLs() string {
	buff := make([]string, 0)
	for _, nurl := range ncd.urls {
		buff = append(buff, fmt.Sprintf("%s://%s:%d", nurl.Scheme, nurl.Fqdn, nurl.Port))
	}
	return strings.Join(buff, ", ")
}

// Connect to the cluster
func (ncd *Ncd) connect() {
	var err error
	log.Print("Connecting...")
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
func (ncd *Ncd) Disconnect() {
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

func (ncd *Ncd) GetPublisher() *nats.Conn {
	return ncd.ncp
}
func (ncd *Ncd) GetSubscriber() *nats.Conn {
	return ncd.ncs
}

// Start starts the Node Controller
func (ncd *Ncd) Start() {
	log.Print("Starting ncd...")
	ncd.connect()
}
