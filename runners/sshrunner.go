/*
	SSH runner.
	Run states remotely without a client over SSH connection.
*/

package runners

import (
	"github.com/isbm/uyuni-ncd/nanostate"
	"os/user"
	"path"
)

type SSHRunner struct {
	_response  *RunnerResponse
	_errcode   int
	_hosts     []string
	_rsapath   string
	_sshport   int
	_sshverify bool
}

func NewSSHRunner() *SSHRunner {
	shr := new(SSHRunner)
	shr._errcode = ERR_INIT
	shr._response = &RunnerResponse{}
	shr._hosts = make([]string, 0)
	shr._sshport = 22
	shr._sshverify = true

	user, _ := user.Current()
	shr._rsapath = path.Join(user.HomeDir, ".ssh")

	return shr
}

// AddHost appends another remote host
func (shr *SSHRunner) AddHost(fqdn string) *SSHRunner {
	shr._hosts = append(shr._hosts, fqdn)
	return shr
}

// SetRSAKeys will set a root directory to the RSA keypair and "known_hosts" database file.
// If an empty string is provided, "$HOME/.ssh" is used instead.
func (shr *SSHRunner) SetRSAKeys(rsapath string) *SSHRunner {
	if rsapath != "" {
		shr._rsapath = rsapath
	}
	return shr
}

// SetSSHHostVerification enables (true, default) or disables (false) the remote host verification,
// based on the "known_hosts" database.
func (shr *SSHRunner) SetSSHHostVerification(hvf bool) *SSHRunner {
	shr._sshverify = hvf
	return shr
}

// SetSSHPort sets an alternative SSH port if needed. Default is 22.
func (shr *SSHRunner) SetSSHPort(port int) *SSHRunner {
	shr._sshport = port
	return shr
}

// Run the compiled and loaded nanostate
func (shr *SSHRunner) Run(state *nanostate.Nanostate) bool {
	errors := 0
	shr._response.Id = state.Id
	shr._response.Description = state.Descr
	groups := make(map[string]RunnerResponseGroup)

	for _, group := range state.Groups {
		resp := &RunnerResponseGroup{
			Errcode: -1,
		}
		response, err := shr.runGroup(group.Group)
		if err != nil {
			resp.Errmsg = err.Error()
			errors++
		} else {
			resp.Response = response
		}
		groups[group.Id] = *resp
	}
	shr._response.Groups = groups

	return errors == 0
}

// Run group of modules
func (shr *SSHRunner) runGroup(group []*nanostate.StateModule) ([]RunnerResponseModule, error) {
	resp := make([]RunnerResponseModule, 0)
	for _, smod := range group {
		cycle := &RunnerResponseModule{
			Module: smod.Module,
		}
		response, err := shr.runModule(smod.Instructions)
		if err != nil {
			cycle.Errcode = ERR_FAILED
			cycle.Errmsg = err.Error()
		} else {
			cycle.Errcode = ERR_OK
			cycle.Response = response
		}
		resp = append(resp, *cycle)
	}
	return resp, nil
}

// Run module with the parameters
func (shr *SSHRunner) runModule(args interface{}) ([]RunnerHostResult, error) {
	result := make([]RunnerHostResult, 0)
	for _, fqdn := range shr._hosts {
		ret := shr.callHost(fqdn, args)
		result = append(result, *ret)
	}
	return result, nil
}

// Call a single host with a series of serial, synchronous commands, ensuring their order.
func (shr *SSHRunner) callHost(fqdn string, args interface{}) *RunnerHostResult {
	response := make(map[string]RunnerStdResult)
	result := &RunnerHostResult{
		Host:     fqdn,
		Response: response,
	}
	for _, command := range args.([]interface{}) {
		for cid, cmd := range command.(map[interface{}]interface{}) {
			remote := NewSshShell(shr._rsapath).SetFQDN(fqdn).SetPort(shr._sshport).SetHostVerification(shr._sshverify).Connect()
			defer remote.Disconnect()
			session := remote.NewSession()
			_, err := session.Run(cmd.(string))
			out := &RunnerStdResult{
				Stdout: session.Outbuff.String(),
				Stderr: session.Errbuff.String(),
			}
			if err != nil {
				out.Errmsg = err.Error()
				out.Errcode = ERR_FAILED
			}
			response[cid.(string)] = *out
		}
	}
	return result
}

// Response returns a map of string/any structure for further processing
func (shr *SSHRunner) Response() *RunnerResponse {
	return shr._response
}

// Errcode returns an error code of the runner
func (shr *SSHRunner) Errcode() int {
	return shr._errcode
}
