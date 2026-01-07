/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package agent

import (
	"fmt"
	"reflect"
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

type ConfigAgentError struct {
	Message string
	Reason  string
}

func (e *ConfigAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to get agent config: %s %s", e.Message, e.Reason))
}

type DescribeAgentError struct {
	Message string
	Reason  string
}

func (e *DescribeAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to get agent description: %s %s", e.Message, e.Reason))
}

type EditAgentError struct {
	Message string
	Reason  string
}

func (e *EditAgentError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Failed to edit agent: %s %s", e.Message, e.Reason))
}

func convertError(err error, to any) error {
	if err == nil {
		return nil
	}
	if to == nil {
		return err
	}

	v := reflect.ValueOf(err)
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return err
	}

	fields := []string{"Message", "Reason"}
	for _, field := range fields {
		f := elem.FieldByName(field)
		if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.String {
			return err
		}

		toField := reflect.ValueOf(to).Elem().FieldByName(field)
		if !toField.IsValid() || !toField.CanSet() || toField.Kind() != reflect.String {
			return err
		}

		toField.SetString(f.String())
	}

	if e, ok := to.(error); ok {
		return e
	}

	return err
}
