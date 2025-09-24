/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"fmt"
	"net/url"
)

func (a *agent) downloadConfig(service, rid string) ([]byte, error) {
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
