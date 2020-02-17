// To-be a plugin in a future.
//
// Event mappers are "generic message to specific instruction" for MQ publisher
/*

# Protecting From Message/Action Reflections

Cluster design relies on only one Leader Node, which receives client changes and
then reflects that to other nodes. The other nodes however, at that time should be
strictly in passive mode. This way they are allowed to only write and do not send
any database changes to anywhere. That means, any changes (input) to them won't
reflect on the rest of the cluster.
*/
package eventmappers

import (
	"crypto/tls"
	"fmt"
	"github.com/isbm/uyuni-ncd/transport"
	"github.com/kolo/xmlrpc"
	"log"
	"net/http"
	"path"
	"reflect"
)

// Used to convert in-messages from Uyuni server to out for cluster
type UyuniEventMapper struct {
	_rpc     *xmlrpc.Client
	_url     string
	_tls     bool
	_user    string
	_pwd     string
	_session string
	intmap   *UyuniIntMap
	actmap   *UyuniActionsMap
	index    map[string]interface{} // For now, at the beginning.
	// Then should be its own type.
	// Used to know what is at boot time in Uyuni
}

func NewUyuniEventMapper() *UyuniEventMapper {
	uem := new(UyuniEventMapper)
	uem._tls = true
	uem.index = make(map[string]interface{})
	uem.intmap = NewUyuniIntMap(uem)
	uem.actmap = NewUyuniActionsMap(uem)
	return uem
}

// Label return a string of a type
func (uem *UyuniEventMapper) Label() string {
	valueOf := reflect.ValueOf(uem)
	if valueOf.Type().Kind() == reflect.Ptr {
		return reflect.Indirect(valueOf).Type().Name()
	} else {
		return valueOf.Type().Name()
	}
}

func (uem *UyuniEventMapper) TopicRoot() string {
	return "/uyuni"
}

// OnReceive tells what to do, once message came from the MQ bus
func (uem *UyuniEventMapper) OnMQReceive(m *ncdtransport.MqMessage) {
	fmt.Println("Uyuni mapper received message:", m.Topic)
	if err := uem.actmap.OnTopic(m); err != nil {
		fmt.Println(err)
	}
}

// Set XML-RPC user
func (uem *UyuniEventMapper) SetRPCUser(user string) *UyuniEventMapper {
	uem._user = user
	return uem
}

// Set XML-RPC password
func (uem *UyuniEventMapper) SetRPCPassword(pwd string) *UyuniEventMapper {
	uem._pwd = pwd
	return uem
}

// SetRPCUrl set URL for the connection point
func (uem *UyuniEventMapper) SetRPCUrl(url string) *UyuniEventMapper {
	uem._url = url
	return uem
}

// SetTLSVerify is to verify certs on SSL/TLS connections
func (uem *UyuniEventMapper) SetTLSVerify(verify bool) *UyuniEventMapper {
	uem._tls = verify
	return uem
}

// This makes all the required indexes of common Uyuni Server data
func (uem *UyuniEventMapper) IndexCommonData() {
	definitions := map[string]string{
		"channel": "channel.listAllChannels",
	}
	for topic, rpcf := range definitions {
		fmt.Println("Indexing", topic, "...")
		uem.index[topic] = uem.scall(rpcf)
	}
}

// Authenticate to Uyuni server
func (uem *UyuniEventMapper) auth() {
	var err error
	var res interface{}
	if uem._user != "" {
		res, err = uem.call("auth.login", uem._user, uem._pwd)
		if err != nil {
			log.Fatal("Login error:", err.Error())
		}
		uem._session = res.(string)
		log.Println("AUTH: Session:", uem._session)
	} else {
		log.Fatalf("User needs to be defined for XML-RPC login")
	}
}

// Internall sessioned call for the XML-RPC
func (uem *UyuniEventMapper) scall(function string, args ...interface{}) interface{} {
	var res interface{}

	if uem._session == "" {
		log.Println("SCALL has no session:", uem._session)
		uem.auth()
	}

	_args := []interface{}{uem._session}
	_args = append(_args, args...)

	res, err := uem.call(function, _args...)
	recall := err != nil
	if err != nil {
		log.Println(err.Error())
	}
	if recall {
		res, err = uem.call(function, _args...)
		if err != nil {
			log.Fatalln("XML-RPC crash:", err.Error())
		}
	}

	return res
}

// Internall direct call for the XML-RPC (raw)
func (uem *UyuniEventMapper) call(function string, args ...interface{}) (interface{}, error) {
	var res interface{}
	return res, uem.GetRpc().Call(function, args, &res)
}

// Get XML-RPC client connection
func (uem *UyuniEventMapper) GetRpc() *xmlrpc.Client {
	if uem._rpc == nil {
		if uem._url == "" {
			panic("XML-RPC client needs an URL to connect")
		}
		var err error
		uem._rpc, err = xmlrpc.NewClient(uem._url, &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: uem._tls},
		})
		if err != nil {
			log.Fatalf("Unable to connect to %s: %s", uem._url, err.Error())
		}
	}
	return uem._rpc
}

// OnIntReceive converts a sendable to everyone mesage about what just happened at Uyuni Server
func (uem *UyuniEventMapper) OnIntReceive(m *ncdtransport.InternalEventMessage) *ncdtransport.MqMessage {
	msg := ncdtransport.NewMqMessage()
	msg.Action = m.Action

	payload, err := uem.intmap.OnTopic(m)
	if err != nil {
		fmt.Println("No actions defined on table", m.Topic)
	} else {
		msg.Topic = path.Join(uem.TopicRoot(), m.Topic)
		msg.Payload = payload
	}

	return msg
}
