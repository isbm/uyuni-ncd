package runners

type RunnerStdResult struct {
	Stdout  string
	Stderr  string
	Errmsg  string
	Errcode int
}

type RunnerHostResult struct {
	Host     string
	Response map[string]RunnerStdResult
}

type RunnerResponseModule struct {
	Module   string
	Errcode  int
	Errmsg   string
	Response []RunnerHostResult
}
type RunnerResponseGroup struct {
	Errcode  int
	Errmsg   string
	Response []RunnerResponseModule
}

type RunnerResponse struct {
	Id          string
	Description string
	Groups      map[string]RunnerResponseGroup
}

func (rr *RunnerResponse) Serialise() map[interface{}]interface{} {
	return nil
}
