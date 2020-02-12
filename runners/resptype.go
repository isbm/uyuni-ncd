package runners

type RunnerResponse struct {
	Id          string
	Description string
	Groups      map[string][]struct {
		Module   string
		Errcode  int
		Errmsg   string
		Response []struct {
			Results []struct {
				Host     string
				Response map[string]string
			}
		}
	}
}
