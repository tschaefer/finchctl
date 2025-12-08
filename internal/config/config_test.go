/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package config

import (
	"os"
	"os/user"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

type stack struct {
	Hostname string
	Username string
	Password string
}

func Test_ErrorString(t *testing.T) {
	err := &ConfigError{
		Message: "test message",
		Reason:  "test reason",
	}

	wanted := "Config error: test message test reason"
	assert.Equal(t, wanted, err.Error(), "error message")
}

func Test_UpdateStackAuthFailIfPermissionDenied(t *testing.T) {
	user, err := user.Current()
	assert.NoError(t, err, "get current user")

	if user.Uid == "0" {
		t.Skip("skipping permission denied test as current user is root")
	}

	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	err = os.Chmod(cfgLoc, 0000)
	assert.NoError(t, err, "change permissions of config directory")

	stack := newStack()
	err = UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	wanted := "Config error: open " + cfgLoc + "/config.json: permission denied"
	assert.EqualError(t, err, wanted, "update stack")
}

func Test_UpdateStackAuthCreateConfigIfNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	assert.NoError(t, err, "update stack")

	_, err = os.Stat(cfgLoc + "/config.json")
	assert.False(t, os.IsNotExist(err), "create config file")
}

func Test_LookupStackAuthReturnErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	assert.NoError(t, err, "update stack")

	hostname := gofakeit.DomainName()
	_, err = LookupStackAuth(hostname)
	wanted := "Config error: stack not found"
	assert.EqualError(t, err, wanted, "lookup stack")
}

func Test_LookupStackAuthReturnCredentialsIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	assert.NoError(t, err, "update stack")

	token, err := LookupStackAuth(stack.Hostname)
	assert.NoError(t, err, "lookup stack")

	assert.Equal(t, encodeToken(stack.Username, stack.Password), token, "auth token")
}

func Test_RemoveStackAuthReturnNoErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	assert.NoError(t, err, "update stack")

	hostname := gofakeit.DomainName()

	err = RemoveStackAuth(hostname)
	assert.NoError(t, err, "remove stack")
}

func Test_RemoveStackAuthSucceedIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	assert.NoError(t, err, "update stack")

	err = RemoveStackAuth(stack.Hostname)
	assert.NoError(t, err, "remove stack")

	_, err = LookupStackAuth(stack.Hostname)
	assert.Error(t, err, "lookup stack")

	wanted := "Config error: stack not found"
	assert.Equal(t, wanted, err.Error(), "error message")
}

func setup(t *testing.T) string {
	cfgLoc, err := os.MkdirTemp("", "finch-test")
	assert.NoError(t, err, "create temp dir for config")

	err = os.Setenv(ConfigLocationEnv, cfgLoc)
	assert.NoError(t, err, "set config location env var")

	return cfgLoc
}

func teardown(cfgLoc string, t *testing.T) {
	err := os.Unsetenv(ConfigLocationEnv)
	assert.NoError(t, err, "unset config location env var")

	err = os.RemoveAll(cfgLoc)
	assert.NoError(t, err, "remove temp config dir")
}

func newStack() stack {
	return stack{
		Hostname: gofakeit.DomainName(),
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 12),
	}
}
