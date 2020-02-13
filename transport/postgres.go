package eventbus

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"os/user"
	"time"
)

type PgEventCallback func(payload interface{})

type PgEventListener struct {
	_host      string
	_port      int
	_sslmode   bool
	_dbname    string
	_user      string
	_password  string
	_channel   string
	_callbacks []PgEventCallback
}

func NewPgEventListener() *PgEventListener {
	pel := new(PgEventListener)
	pel._port = 5432
	pel._sslmode = true
	pel._host = "localhost"
	pel._dbname = "postgres"
	pel._callbacks = make([]PgEventCallback, 0)

	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	pel._user = u.Username

	return pel
}

// Add a callback on event
func (pel *PgEventListener) AddCallback(callback PgEventCallback) *PgEventListener {
	pel._callbacks = append(pel._callbacks, callback)
	return pel
}

// SetChannel sets a listening channel name
func (pel *PgEventListener) SetChannel(channel string) *PgEventListener {
	pel._channel = channel
	return pel
}

// SetHost changes hostname from "localhost" to whatever else.
func (pel *PgEventListener) SetHost(host string) *PgEventListener {
	pel._host = host
	return pel
}

// SetPort another than 5432 port
func (pel *PgEventListener) SetPort(port int) *PgEventListener {
	pel._port = port
	return pel
}

// SetSSLMode turns ON or OFF the SSL connection. Some setups doesn't support TLS connections.
func (pel *PgEventListener) SetSSLMode(mode bool) *PgEventListener {
	pel._sslmode = mode
	return pel
}

// SetDBName sets the name of the database. Default is "postgres"
func (pel *PgEventListener) SetDBName(name string) *PgEventListener {
	pel._dbname = name
	return pel
}

// SetUser sets a user name. Default is current user
func (pel *PgEventListener) SetUser(user string) *PgEventListener {
	pel._user = user
	return pel
}

// SetPassword
func (pel *PgEventListener) SetPassword(password string) *PgEventListener {
	pel._password = password
	return pel
}

// Notification monitor
func (pel *PgEventListener) notificationMonitor(listener *pq.Listener) {
	for {
		select {
		case ch := <-listener.Notify:
			var payload interface{}
			if err := json.Unmarshal([]byte(ch.Extra), &payload); err != nil {
				fmt.Println("Error getting JSON:", err.Error()) // XXX: Logger!!
			}
			for _, callback := range pel._callbacks {
				go callback(payload)
			}
			return
		case <-time.After(10 * time.Second):
			go func() {
				if err := listener.Ping(); err != nil {
					panic(err)
				}
			}()
			return
		}
	}
}

// Format connection string
func (pel *PgEventListener) getConnString() string {
	var sslmode string
	if pel._sslmode {
		sslmode = "enable"
	} else {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		pel._host, pel._port, pel._dbname, pel._user, pel._password, sslmode)
}

// Log errors during the listening
func (pel *PgEventListener) errorLogger(event pq.ListenerEventType, err error) {
	if err != nil {
		fmt.Println("Error during listening:", err.Error()) // XXX: Logger!
	}
}

func (pel *PgEventListener) Start() {
	if pel._channel == "" {
		panic(errors.New("Channel is missing"))
	}
	_, err := sql.Open("postgres", pel.getConnString())
	if err != nil {
		panic(err)
	}
	listener := pq.NewListener(pel.getConnString(), 10*time.Second, time.Minute, pel.errorLogger)
	if err := listener.Listen(pel._channel); err != nil {
		panic(err)
	}
	for {
		pel.notificationMonitor(listener)
	}
}
