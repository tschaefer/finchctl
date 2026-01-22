/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

type ConfigError struct {
	Message string
	Reason  string
}

func (e *ConfigError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("Config error: %s %s", e.Message, e.Reason))
}

const (
	ConfigLocationEnv string = "FINCH_CONFIG"
)

type Stacks struct {
	List []Stack `json:"stacks"`
}

type Stack struct {
	Name string `json:"name,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

func UpdateStack(name string, certPEM, keyPEM []byte) error {
	var stacks Stacks
	if exist() {
		if err := backup(); err != nil {
			return &ConfigError{Message: "failed to backup config", Reason: err.Error()}
		}

		cfg, err := read()
		if err != nil {
			return &ConfigError{Message: "failed to read config", Reason: err.Error()}
		}
		stacks = *cfg
	}

	exists := false
	for _, stack := range stacks.List {
		if stack.Name == name {
			exists = true
			break
		}
	}

	if exists {
		stacks.List = slices.DeleteFunc(stacks.List, func(s Stack) bool {
			return s.Name == name
		})
	}

	stacks.List = append(stacks.List, Stack{
		Name: name,
		Cert: base64.StdEncoding.EncodeToString(certPEM),
		Key:  base64.StdEncoding.EncodeToString(keyPEM),
	})

	return write(&stacks)
}

func LookupStack(name string) ([]byte, []byte, error) {
	if !exist() {
		return nil, nil, &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	stacks, err := read()
	if err != nil {
		return nil, nil, &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	for _, stack := range stacks.List {
		if stack.Name == name {
			certPEM, err := base64.StdEncoding.DecodeString(stack.Cert)
			if err != nil {
				return nil, nil, &ConfigError{Message: "failed to decode certificate", Reason: err.Error()}
			}
			keyPEM, err := base64.StdEncoding.DecodeString(stack.Key)
			if err != nil {
				return nil, nil, &ConfigError{Message: "failed to decode key", Reason: err.Error()}
			}
			return certPEM, keyPEM, nil
		}
	}

	return nil, nil, &ConfigError{Message: "stack not found", Reason: ""}
}

func RemoveStack(name string) error {
	if !exist() {
		return &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	stacks, err := read()
	if err != nil {
		return &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	index := -1
	for i, stack := range stacks.List {
		if stack.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	}

	if err := backup(); err != nil {
		return &ConfigError{Message: "failed to backup config", Reason: err.Error()}
	}

	stacks.List = slices.Delete(stacks.List, index, index+1)

	return write(stacks)
}

func ListStacks() ([]string, error) {
	if !exist() {
		return nil, &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	stacks, err := read()
	if err != nil {
		return nil, &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	var names []string
	for _, stack := range stacks.List {
		names = append(names, stack.Name)
	}

	return names, nil
}

func backup() error {
	data, err := os.ReadFile(path())
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s~", path()), data, 0400)
	if err != nil {
		return err
	}

	return nil
}

func write(stacks *Stacks) error {
	data, err := json.MarshalIndent(stacks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path(), data, 0600)
}

func read() (*Stacks, error) {
	data, err := os.ReadFile(path())
	if err != nil {
		return nil, err
	}

	var stacks Stacks
	if err := json.Unmarshal(data, &stacks); err != nil {
		return nil, err
	}

	return &stacks, nil
}

func path() string {
	var dir string
	dir = os.Getenv(ConfigLocationEnv)

	if dir == "" {
		var err error
		dir, err = os.UserHomeDir()
		if err != nil {
			panic(&ConfigError{Message: err.Error(), Reason: ""})
		}
		dir = fmt.Sprintf("%s/.config", dir)
	}

	return fmt.Sprintf("%s/finch.json", dir)
}

func exist() bool {
	if _, err := os.Stat(path()); err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
