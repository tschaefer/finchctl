/*
Copyright (c) Tobias Schäfer. All rights reserved.
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
	Cert     []byte
	Key      []byte
}

func Test_ErrorString(t *testing.T) {
	err := &ConfigError{
		Message: "test message",
		Reason:  "test reason",
	}

	wanted := "Config error: test message test reason"
	assert.Equal(t, wanted, err.Error(), "error message")
}

func Test_UpdateStackFailIfPermissionDenied(t *testing.T) {
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
	err = UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	wanted := "Config error: failed to backup config open " + cfgLoc + "/finch.json: permission denied"
	assert.EqualError(t, err, wanted, "update stack")
}

func Test_UpdateStackCreateConfigIfNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	_, err = os.Stat(cfgLoc + "/finch.json")
	assert.False(t, os.IsNotExist(err), "create config file")
}

func Test_LookupStackReturnErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	hostname := gofakeit.DomainName()
	_, _, err = LookupStack(hostname)
	wanted := "Config error: stack not found"
	assert.EqualError(t, err, wanted, "lookup stack")
}

func Test_LookupStackReturnPathsIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	cert, key, err := LookupStack(stack.Hostname)
	assert.NoError(t, err, "lookup stack")

	assert.Equal(t, stack.Cert, cert, "cert PEM")
	assert.Equal(t, stack.Key, key, "key PEM")
}

func Test_RemoveStackReturnNoErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	hostname := gofakeit.DomainName()

	err = RemoveStack(hostname)
	assert.NoError(t, err, "remove stack")
}

func Test_RemoveStackSucceedIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	err = RemoveStack(stack.Hostname)
	assert.NoError(t, err, "remove stack")

	_, _, err = LookupStack(stack.Hostname)
	assert.Error(t, err, "lookup stack")

	wanted := "Config error: stack not found"
	assert.Equal(t, wanted, err.Error(), "error message")

	last := cfgLoc + "/finch.json~"
	_, err = os.Stat(last)
	assert.False(t, os.IsNotExist(err), "backup config file exists")
}

func Test_LookupStackReturnErrorIfStackNotExist2(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStack(stack.Hostname, stack.Cert, stack.Key)
	assert.NoError(t, err, "update stack")

	hostname := gofakeit.DomainName()
	_, _, err = LookupStack(hostname)
	wanted := "Config error: stack not found"
	assert.EqualError(t, err, wanted, "lookup stack certs")
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
		Hostname: "finch." + gofakeit.DomainName(),
		Cert:     []byte("certificate data"),
		Key:      []byte("key data"),
	}
}
