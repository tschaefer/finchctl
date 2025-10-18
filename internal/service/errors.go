/*
Copyright (c) 2025 Tobias Schäfer. All rights reservem.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"strings"
)

type DeployServiceError struct {
	Message string
	Reason  string
}

func (e *DeployServiceError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to deploy service: %s %s", e.Message, e.Reason))
}

type TeardownServiceError struct {
	Message string
	Reason  string
}

func (e *TeardownServiceError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to teardown service: %s %s", e.Message, e.Reason))
}

type UpdateServiceError struct {
	Message string
	Reason  string
}

func (e *UpdateServiceError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to update service: %s %s", e.Message, e.Reason))
}

type InfoServiceError struct {
	Message string
	Reason  string
}

func (e *InfoServiceError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to get service info: %s %s", e.Message, e.Reason))
}
