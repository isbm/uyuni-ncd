package runners

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	knownHostsVerification "golang.org/x/crypto/ssh/knownhosts"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
)

type NodeStage struct {
	_fqdn     string
	_port     int
	_rsa_keys string
	_rsa_priv string
	_rsa_pub  string
	_user     string
	_hkb      ssh.HostKeyCallback
	_conn     *ssh.Client
	_sessions map[string]*NodeSSHSession
	_runner   *SSHRunner
}

// Constructor. Needs to be given a location of SSH keys, including "known_hosts".
// As an example: "/home/someuser/.ssh".
func NewNodeStage(rsa string) *NodeStage {
	ns := new(NodeStage)
	ns._rsa_keys = rsa
	ns._port = 22

	u, err := user.Current()
	if err != nil {
		panic("Unable to obtain current user")
	}
	ns._user = u.Username

	ns.verifyRSAKeyPath()
	ns._rsa_priv = "id_rsa"
	ns._rsa_pub = "id_rsa.pub"

	// Manage sessions
	ns._sessions = make(map[string]*NodeSSHSession)
	ns._runner = NewSSHRunner()

	ns.SetHostVerification(true)

	return ns
}

// SetRemoteUsername sets remote username. Default is the current username.
func (ns *NodeStage) SetRemoteUsername(username string) *NodeStage {
	ns._user = username
	return ns
}

// SetHostVerification to true or false on SSH connection. Default is set to True.
func (ns *NodeStage) SetHostVerification(hv bool) *NodeStage {
	var err error
	if hv {
		khpath := path.Join(ns._rsa_keys, "known_hosts")
		ns._hkb, err = knownHostsVerification.New(khpath)
		if err != nil {
			panic("Attempt to setup secure host verification but 'known_hosts' database file was not found")
		}
	} else {
		ns._hkb = ssh.InsecureIgnoreHostKey()
	}
	return ns
}

// SetFQDN of the node that is going to be staged
func (ns *NodeStage) SetFQDN(fqdn string) *NodeStage {
	ns._fqdn = fqdn
	return ns
}

// SetPort the opened SSH port on the node that is going to be staged. Default is a standard 22.
func (ns *NodeStage) SetPort(port int) *NodeStage {
	ns._port = port
	return ns
}

// SetRSAPrivKey sets a path to the private RSA key for
// the SSH connection.
func (ns *NodeStage) SetRSAPrivKey(name string) *NodeStage {
	ns._rsa_priv = name
	return ns
}

// SetRSAPubKey sets a path to the public RSA key for
// the SSH connection. This key should be deployed
// to the target node.
func (ns *NodeStage) SetRSAPubKey(name string) *NodeStage {
	ns._rsa_pub = name
	return ns
}

// Connect opens an SSH connection to the remote machine
func (ns *NodeStage) Connect() *NodeStage {
	if ns._conn != nil {
		return ns
	}

	var err error
	signer, err := ssh.ParsePrivateKey(ns.getFileContent(path.Join(ns._rsa_keys, ns._rsa_priv)))
	if err != nil {
		panic("ERROR: Unable to parse private RSA key")
	}
	sshconf := &ssh.ClientConfig{
		User: ns._user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ns._hkb,
	}
	ns._conn, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", ns._fqdn, ns._port), sshconf)
	if err != nil {
		log.Println("ERROR: Unable to connect:", err.Error())
	}
	return ns
}

func (ns *NodeStage) NewSession() *NodeSSHSession {
	if ns._conn == nil {
		panic("Attempt to open a new session when no connection has been yet made")
	}
	session := NewNodeSSHSession(ns._conn)
	ns._sessions[session.Id] = session
	return session
}

// Disconnect closes the SSH connection
func (ns *NodeStage) Disconnect() *NodeStage {
	if ns._conn != nil {
		ns._conn.Close()
		ns._conn = nil
	}

	return ns
}

/////////////////////////////// Internal
func (ns *NodeStage) getFileContent(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: Unable to read file '%s'", path)
	}
	return data
}

func (ns *NodeStage) verifyRSAKeyPath() {
	if st, err := os.Stat(ns._rsa_keys); os.IsNotExist(err) || !st.IsDir() {
		panic("Path to RSA files does not exist or is not a directory")
	}
}

//////////////////////////////// API

// Run node staging
func (ns *NodeStage) Run(scenario map[string]interface{}) {
	playbook := Playbook()
	playbook.Load(scenario)
	for _, group := playbook.GetSections() {
		ns._runner.Run(group)
	}
}
