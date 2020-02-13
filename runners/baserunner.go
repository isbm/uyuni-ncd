package runners

import (
	"fmt"
	"github.com/isbm/uyuni-ncd/nanostate"
)

type IBaseRunner interface {
	callShell(args interface{}) ([]RunnerHostResult, error)
}

type BaseRunner struct {
	ref       IBaseRunner
	_response *RunnerResponse
	_errcode  int
}

// Run the compiled and loaded nanostate
func (br *BaseRunner) Run(state *nanostate.Nanostate) bool {
	errors := 0
	br._response.Id = state.Id
	br._response.Description = state.Descr
	groups := make(map[string]RunnerResponseGroup)

	for _, group := range state.Groups {
		resp := &RunnerResponseGroup{
			Errcode: -1,
		}
		response, err := br.runGroup(group.Group)
		if err != nil {
			resp.Errmsg = err.Error()
			errors++
		} else {
			resp.Response = response
		}
		groups[group.Id] = *resp
	}
	br._response.Groups = groups

	return errors == 0
}

// Run group of modules
func (br *BaseRunner) runGroup(group []*nanostate.StateModule) ([]RunnerResponseModule, error) {
	resp := make([]RunnerResponseModule, 0)
	for _, smod := range group {
		cycle := &RunnerResponseModule{
			Module: smod.Module,
		}
		if cycle.Module == "shell" {
			response, err := br.ref.callShell(smod.Instructions)
			if err != nil {
				cycle.Errcode = ERR_FAILED
				cycle.Errmsg = err.Error()
			} else {
				cycle.Errcode = ERR_OK
				cycle.Response = response
			}
			resp = append(resp, *cycle)
		} else {
			fmt.Println(">>> ERROR: module", cycle.Module, "is not supported")
		}
	}
	return resp, nil
}

func (br *BaseRunner) callShell(args interface{}) ([]RunnerHostResult, error) {
	panic("Abstract method")
}

// Response returns a map of string/any structure for further processing
func (br *BaseRunner) Response() *RunnerResponse {
	return br._response
}

// Errcode returns an error code of the runner
func (br *BaseRunner) Errcode() int {
	return br._errcode
}