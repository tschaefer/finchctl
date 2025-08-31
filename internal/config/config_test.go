/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package config

import (
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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

	expected := "Config error: test message test reason"
	if err.Error() != expected {
		t.Fatalf("expected error string %q, got %q", expected, err.Error())
	}
}

func Test_UpdateStackAuthFailIfPermissionDenied(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	if err := os.Chmod(cfgLoc, 0000); err != nil {
		t.Fatalf("failed to change permissions: %v", err)
	}

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	if err == nil {
		t.Fatalf("expected error")
	}

	wanted := "Config error: open " + cfgLoc + "/config.json: permission denied"
	if err.Error() != wanted {
		t.Fatalf("expected error %q, got %q", wanted, err.Error())
	}
}

func Test_UpdateStackAuthCreateConfigIfNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	if err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(cfgLoc + "/config.json"); os.IsNotExist(err) {
		t.Fatalf("expected config file to be created")
	}
}

func Test_LookupStackAuthReturnErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	if err != nil {
		t.Error(err)
	}

	hostname := gofakeit.DomainName()
	_, _, err = LookupStackAuth(hostname)
	if err == nil {
		t.Fatalf("expected error")
	}

	wanted := "Config error: stack not found"
	if err.Error() != wanted {
		t.Fatalf("expected error %q, got %q", wanted, err.Error())
	}
}

func Test_LookupStackAuthReturnCredentialsIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	if err != nil {
		t.Error(err)
	}

	username, password, err := LookupStackAuth(stack.Hostname)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stack.Username != username {
		t.Fatalf("expected username to be '%q', got %q", stack.Username, username)
	}

	if stack.Password != password {
		t.Fatalf("expected password to be '%q', got %q", stack.Password, password)
	}
}

func Test_RemoveStackAuthReturnNoErrorIfStackNotExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	if err != nil {
		t.Error(err)
	}

	hostname := gofakeit.DomainName()

	err = RemoveStackAuth(hostname)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func Test_RemoveStackAuthSucceedIfStackExist(t *testing.T) {
	cfgLoc := setup(t)
	defer teardown(cfgLoc, t)

	stack := newStack()
	err := UpdateStackAuth(stack.Hostname, stack.Username, stack.Password)
	if err != nil {
		t.Error(err)
	}

	err = RemoveStackAuth(stack.Hostname)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, _, err = LookupStackAuth(stack.Hostname)
	if err == nil {
		t.Fatalf("expected error")
	}

	wanted := "Config error: stack not found"
	if err.Error() != wanted {
		t.Fatalf("expected error %q, got %q", wanted, err.Error())
	}
}

func setup(t *testing.T) string {
	cfgLoc, err := os.MkdirTemp("", "finch-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv(ConfigLocationEnv, cfgLoc); err != nil {
		t.Fatal(err)
	}

	return cfgLoc
}

func teardown(cfgLoc string, t *testing.T) {
	if err := os.Unsetenv(ConfigLocationEnv); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(cfgLoc); err != nil {
		t.Error(err)
	}
}

func newStack() stack {
	return stack{
		Hostname: gofakeit.DomainName(),
		Username: gofakeit.Username(),
		Password: gofakeit.Password(true, true, true, true, false, 12),
	}
}
