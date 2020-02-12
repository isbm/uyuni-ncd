package runners

import (
	"github.com/isbm/uyuni-ncd/nanostate"
)

type LocalRunner struct {
	_response map[string]interface{}
	_errcode  int
}

func NewLocalRunner() *LocalRunner {
	lr := new(LocalRunner)
	lr._errcode = ERR_INIT
	return lr
}

// Run the compiled and loaded nanostate
func (lr *LocalRunner) Run(state *nanostate.Nanostate) error {
	return nil
}

// Response returns a map of string/any structure for further processing
func (lr *LocalRunner) Response() map[string]interface{} {
	return lr._response
}

// Errcode returns an error code of the runner
func (lr *LocalRunner) Errcode() int {
	return lr._errcode
}
