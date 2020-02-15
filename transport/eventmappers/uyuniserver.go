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
	index    map[string]interface{} // For now, at the beginning.
	// Then should be its own type.
	// Used to know what is at boot time in Uyuni
}

func NewUyuniEventMapper() *UyuniEventMapper {
	uem := new(UyuniEventMapper)
	uem._tls = true
	uem.index = make(map[string]interface{})
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
func (uem *UyuniEventMapper) OnReceive(m ncdtransport.MqMessage) {
	fmt.Println("Uyuni mapper received message:", m.Topic)
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
	} else {
		log.Fatalf("User needs to be defined for XML-RPC login")
	}
}

// Internall sessioned call for the XML-RPC
func (uem *UyuniEventMapper) scall(function string, args ...interface{}) interface{} {
	var res interface{}
	recall := false
	if uem._session == "" {
		uem.auth()
	}
	res, err := uem.call(function, args)
	if err != nil {
		uem.auth() // It might be token expiration, so try again. Unless Uyuni changes to a better error handling...
		recall = true
	}
	if recall {
		res, err = uem.call(function, "", args)
		if err != nil {
			log.Println("Function:", function, "Args:", args)
			log.Fatal("Second pass failure:", err) // OK, now fail
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

// UyuniEventToMessage tells to everyone through the message bus what just happened at Uyuni Server
func (uem *UyuniEventMapper) UyuniEventToMessage(m *ncdtransport.DbEventMessage) *ncdtransport.MqMessage {
	msg := ncdtransport.NewMqMessage()
	switch m.Table {
	case "rhnchannel":
		switch m.Action {
		case "insert":
			msg.Action = m.Action
			msg.Topic = path.Join(uem.TopicRoot(), "channel")
			msg.Payload = uem.scall("channel.software.getDetails")

		case "update":
		case "delete":
		default:
			fmt.Println("No destination defined on action", m.Action)
		}
	default:
		fmt.Println("No actions defined on table", m.Table)
	}
	return msg
}
