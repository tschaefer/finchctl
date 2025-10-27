/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type RegisterData struct {
	Hostname   string   `json:"hostname"`
	LogSources []string `json:"log_sources"`
	Metrics    bool     `json:"metrics"`
	Profiles   bool     `json:"profiles"`
	Tags       []string `json:"tags"`
}

func (a *agent) __registerDo(service string, data *RegisterData) (string, error) {
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = service
	url.Path = "/finch/api/v1/agent"

	payload, err := json.Marshal(data)
	if err != nil {
		return "", &RegisterAgentError{Message: err.Error(), Reason: ""}
	}

	payload, err = a.target.Request("POST", url, payload)
	if err != nil {
		return "", &RegisterAgentError{Message: err.Error(), Reason: ""}
	}

	var info map[string]string
	if err := json.Unmarshal(payload, &info); err != nil {
		return "", &RegisterAgentError{Message: err.Error(), Reason: "Failed to parse response"}
	}

	return info["rid"], nil
}

func (a *agent) __registerGetAgentConfig(service, rid string) ([]byte, error) {
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = service
	url.Path = fmt.Sprintf("/finch/api/v1/agent/%s/config", rid)

	payload, err := a.target.Request("GET", url, nil)
	if err != nil {
		return nil, &RegisterAgentError{Message: err.Error(), Reason: ""}
	}

	return payload, nil
}

func (a *agent) registerAgent(service string, data *RegisterData) ([]byte, error) {
	rid, err := a.__registerDo(service, data)
	if err != nil {
		return nil, err
	}

	config, err := a.__registerGetAgentConfig(service, rid)
	if err != nil {
		return nil, err
	}

	return config, nil
}
