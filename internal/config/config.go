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
	"path/filepath"
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

	stacks.List = slices.DeleteFunc(stacks.List, func(s Stack) bool {
		return s.Name == name
	})

	stacks.List = append(stacks.List, Stack{
		Name: name,
		Cert: base64.StdEncoding.EncodeToString(certPEM),
		Key:  base64.StdEncoding.EncodeToString(keyPEM),
	})

	return write(&stacks)
}

func LookupStack(name string) (*Stack, error) {
	if !exist() {
		return nil, &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	stacks, err := read()
	if err != nil {
		return nil, &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	for _, stack := range stacks.List {
		if stack.Name == name {
			certPEM, err := base64.StdEncoding.DecodeString(stack.Cert)
			if err != nil {
				return nil, &ConfigError{Message: "failed to decode certificate", Reason: err.Error()}
			}
			keyPEM, err := base64.StdEncoding.DecodeString(stack.Key)
			if err != nil {
				return nil, &ConfigError{Message: "failed to decode key", Reason: err.Error()}
			}
			return &Stack{
				Name: stack.Name,
				Cert: string(certPEM),
				Key:  string(keyPEM),
			}, nil
		}
	}

	return nil, &ConfigError{Message: "stack not found", Reason: ""}
}

func RemoveStack(name string) error {
	if !exist() {
		return &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	stacks, err := read()
	if err != nil {
		return &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	index := slices.IndexFunc(stacks.List, func(s Stack) bool {
		return s.Name == name
	})

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
	p := path()
	data, err := os.ReadFile(p)
	if err != nil {
		return err
	}

	return os.WriteFile(p+"~", data, 0600)
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
	dir := os.Getenv(ConfigLocationEnv)

	if dir == "" {
		var err error
		dir, err = os.UserHomeDir()
		if err != nil {
			panic(&ConfigError{Message: err.Error(), Reason: ""})
		}
		dir = filepath.Join(dir, ".config")
	}

	return filepath.Join(dir, "finch.json")
}

func exist() bool {
	if _, err := os.Stat(path()); err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
