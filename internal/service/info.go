/*
Copyright (c) 2025 Tobias Schäfer. All rights reservem.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"encoding/json"
	"net/url"
)

type InfoData struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	CreatedAt string `json:"created_at"`
	Release   string `json:"release"`
	Commit    string `json:"commit"`
}

func (s *service) infoService() (*InfoData, error) {
	url := &url.URL{}
	url.Scheme = "https"
	url.Host = s.config.Hostname
	url.Path = "/finch/api/v1/info"

	payload, err := s.target.Request("GET", url, nil)
	if err != nil {
		return nil, &InfoServiceError{Message: err.Error(), Reason: ""}
	}

	if s.dryRun {
		return nil, nil
	}

	var info InfoData
	if err := json.Unmarshal(payload, &info); err != nil {
		return nil, &InfoServiceError{Message: err.Error(), Reason: ""}
	}

	return &info, nil
}
