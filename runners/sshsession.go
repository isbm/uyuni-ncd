package runners

import (
	"bytes"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"log"
)

type NodeSSHSession struct {
	Session *ssh.Session
	Id      string
	Outbuff bytes.Buffer
	Errbuff bytes.Buffer
}

func NewNodeSSHSession(ctx *ssh.Client) *NodeSSHSession {
	ss := new(NodeSSHSession)

	uuid, _ := uuid.NewUUID()
	ss.Id = uuid.String()

	var err error
	ss.Session, err = ctx.NewSession()
	if err != nil {
		log.Fatalln("unable to create SSH session:", err.Error())
	}

	ss.Session.Stdout = &ss.Outbuff
	ss.Session.Stderr = &ss.Errbuff

	return ss
}

// Run a command
func (ss *NodeSSHSession) Run(cmd string) (string, error) {
	defer ss.Session.Close()
	err := ss.Session.Run(cmd)
	return ss.Outbuff.String(), err
}
