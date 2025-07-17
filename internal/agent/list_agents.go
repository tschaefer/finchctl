package agent

import (
	"encoding/json"
	"net/url"
)

type ListData struct {
	Hostname   string `json:"hostname"`
	ResourceID string `json:"rid"`
}

func (a *agent) listAgents(service string) (*[]ListData, error) {
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = service
	url.Path = "/finch/api/v1/agent"

	payload, err := a.target.Request("GET", url, nil)
	if err != nil {
		return nil, &ListAgentsError{Message: err.Error(), Reason: ""}
	}

	var list []ListData
	if err := json.Unmarshal(payload, &list); err != nil {
		return nil, &ListAgentsError{Message: err.Error(), Reason: ""}
	}

	return &list, nil
}
