package runners

import (
	"github.com/isbm/uyuni-ncd/nanostate"
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
