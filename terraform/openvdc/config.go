// + build terraform

package openvdc

type config struct {
	apiEndpoint string
}

func (c *config) getApiEndpoint() string {
	if c.apiEndpoint != "" {
		return c.apiEndpoint
	}
  return "0.0.0.0:5000"
}
