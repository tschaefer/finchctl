/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
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
	return fmt.Sprintf("Config error: %s %s", e.Message, e.Reason)
}

const (
	ConfigLocationEnv string = "FINCH_CONFIG"
)

type Config struct {
	Stacks []Stack `json:"stacks"`
}

type Stack struct {
	Name  string `yaml:"name,omitempty"`
	Token string `yaml:"token,omitempty"`
}

func UpdateStackAuth(name, username, password string) error {
	var config Config
	if fileExists() {
		cfg, err := ReadConfig()
		if err != nil {
			return &ConfigError{Message: err.Error(), Reason: ""}
		}
		config = *cfg
	}

	exists := false
	for _, authConfig := range config.Stacks {
		if authConfig.Name == name {
			exists = true
			break
		}
	}

	if exists {
		config.Stacks = slices.DeleteFunc(config.Stacks, func(s Stack) bool {
			return s.Name == name
		})
	}

	config.Stacks = append(config.Stacks, Stack{
		Name:  name,
		Token: encodeToken(username, password),
	})

	return WriteConfig(&config)
}

func LookupStackAuth(name string) (string, string, error) {
	if !fileExists() {
		return "", "", &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	config, err := ReadConfig()
	if err != nil {
		return "", "", &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	var token string
	for _, authConfig := range config.Stacks {
		if authConfig.Name == name {
			token = authConfig.Token
			break
		}
	}
	if token == "" {
		return "", "", &ConfigError{Message: "stack not found", Reason: ""}
	}

	username, password, err := decodeToken(token)
	if err != nil {
		return "", "", &ConfigError{Message: "failed to decode auth token", Reason: err.Error()}
	}

	return username, password, nil
}

func RemoveStackAuth(name string) error {
	if !fileExists() {
		return &ConfigError{Message: "config file does not exist", Reason: ""}
	}

	config, err := ReadConfig()
	if err != nil {
		return &ConfigError{Message: "failed to read config", Reason: err.Error()}
	}

	index := -1
	for i, authConfig := range config.Stacks {
		if authConfig.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		return nil
	}

	config.Stacks = slices.Delete(config.Stacks, index, index+1)

	return WriteConfig(config)
}

func fileExists() bool {
	configFile := configFile()

	if _, err := os.Stat(configFile); err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func WriteConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return &ConfigError{Message: err.Error(), Reason: ""}
	}

	return os.WriteFile(configFile(), data, 0600)
}

func ReadConfig() (*Config, error) {
	configFile := configFile()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, &ConfigError{Message: err.Error(), Reason: ""}
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, &ConfigError{Message: err.Error(), Reason: ""}
	}

	return &config, nil
}

func configFile() string {
	var dir string
	dir = os.Getenv(ConfigLocationEnv)

	if dir == "" {
		var err error
		dir, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dir = fmt.Sprintf("%s/.finch", dir)
	}

	return fmt.Sprintf("%s/config.json", dir)
}

func decodeToken(input string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", "", &ConfigError{Message: err.Error(), Reason: ""}
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", &ConfigError{Message: "invalid token format", Reason: ""}
	}

	return parts[0], parts[1], nil
}

func encodeToken(username, password string) string {
	plain := fmt.Sprintf("%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(plain))

	return encoded
}
