package runners

import (
	"github.com/isbm/uyuni-ncd/nanostate"
)

const (
	ERR_OK = 0
	ERR_FAILED = 1
	ERR_TIMEOUT = 2 // Prepared, but unprocessed
	ERR_INIT = 255
)

// Interface for the runner
type Runner interface {
	// Run the compiled and loaded nanostate
	Run(state *nanostate.Nanostate) error

	// Response returns a map of string/any structure for further processing
	Response() map[string]interface{}

	// Errcode returns an error code of the runner
	Errcode() int
}
