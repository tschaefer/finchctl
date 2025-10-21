package agent

import (
	"fmt"
	"strings"
)

type DeployAgentError struct {
	Message string
	Reason  string
}

func (e *DeployAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to deploy agent: %s %s", e.Message, e.Reason))
}

type RegisterAgentError struct {
	Message string
	Reason  string
}

func (e *RegisterAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to register agent: %s %s", e.Message, e.Reason))
}

type TeardownAgentError struct {
	Message string
	Reason  string
}

func (e *TeardownAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to teardown agent: %s %s", e.Message, e.Reason))
}

type ListAgentsError struct {
	Message string
	Reason  string
}

func (e *ListAgentsError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to list agents: %s %s", e.Message, e.Reason))
}

type DeregisterAgentError struct {
	Message string
	Reason  string
}

func (e *DeregisterAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to deregister agent: %s %s", e.Message, e.Reason))
}

type UpdateAgentError struct {
	Message string
	Reason  string
}

func (e *UpdateAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to update agent: %s %s", e.Message, e.Reason))
}
