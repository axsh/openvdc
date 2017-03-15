package vm

type Base struct {
}

var SupportedAPICalls = []string{
	"/api.Instance/Start",
	"/api.Instance/Run",
	"/api.Instance/Stop",
	"/api.Instance/Reboot",
	"/api.Instance/Console",
	"/api.Instance/Log",
}

func (*Base) IsSupportAPI(method string) bool {
	for _, m := range SupportedAPICalls {
		if m == method {
			return true
		}
	}
	return false
}
