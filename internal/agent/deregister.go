package agent

import (
	"fmt"
	"net/url"
)

func (a *agent) deregisterAgent(service, resourceID string) error {
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = service
	url.Path = fmt.Sprintf("/finch/api/v1/agent/%s", resourceID)

	_, err := a.target.Request("DELETE", url, nil)
	if err != nil {
		return &DeregisterAgentError{Message: err.Error(), Reason: ""}
	}

	return nil
}
