/*
Copyright (c) 2025 Tobias Schäfer. All rights reservem.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

import (
	"fmt"
	"reflect"
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
