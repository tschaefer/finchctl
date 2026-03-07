/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package service

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tschaefer/finchctl/internal/config"
	"github.com/tschaefer/finchctl/internal/target"
)

type track struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func Test_Deploy(t *testing.T) {
	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
		Config: &ServiceConfig{
			Hostname: "localhost",
		},
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Deploy()
	})
	assert.NoError(t, err, "deploy service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 57, "number of log lines")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Skipping readiness check due to dry-run mode"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
		Config: &ServiceConfig{
			Hostname: "localhost",
		},
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Deploy()
	})
	assert.NoError(t, err, "deploy service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Teardown(t *testing.T) {
	setupAssets(t)
	defer teardownAssets(t)

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Teardown()
	})
	assert.NoError(t, err, "teardown service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 8, "number of log lines mismatch")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo rm -rf /tmp/finch-test-lib-[0-9]+' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Teardown()
	})
	assert.NoError(t, err, "teardown service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Update(t *testing.T) {
	setupAssets(t)
	defer teardownAssets(t)

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Update()
	})
	assert.NoError(t, err, "update service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 54, "number of log lines")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo docker image prune --force' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "update service")

	record = capture(func() {
		err = s.Update()
	})
	assert.NoError(t, err, "update service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_RotateSecret(t *testing.T) {
	setupAssets(t)
	defer teardownAssets(t)

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.RotateSecret()
	})
	assert.NoError(t, err, "rotate secret")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 9, "number of log lines mismatch")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo docker compose --file .+/docker-compose.yaml restart finch' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.RotateSecret()
	})
	assert.NoError(t, err, "rotate secret")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")
}

func Test_RotateCertificate(t *testing.T) {
	setupAssets(t)
	defer teardownAssets(t)

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.RotateCertificate()
	})
	assert.NoError(t, err, "rotate certificate")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 8, "number of log lines mismatch")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'rm -f .+/traefik/etc/certs.d/ca.pem' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.RotateCertificate()
	})
	assert.NoError(t, err, "rotate certificate")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Register(t *testing.T) {
	setupAssets(t)
	cfgDir := os.Getenv(config.ConfigLocationEnv)
	err := os.Unsetenv(config.ConfigLocationEnv)
	assert.NoError(t, err, "unset cfg dir env")
	defer func() {
		err := os.Setenv(config.ConfigLocationEnv, cfgDir)
		assert.NoError(t, err, "restore cfg dir env")
		teardownAssets(t)
	}()

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Register()
	})
	assert.NoError(t, err, "register service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 7, "number of log lines mismatch")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = `Copying from '.+' to '.+/traefik/etc/certs.d/rid:finchctl:.+\.pem' as .+@localhost`
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Register()
	})
	assert.NoError(t, err, "register service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Deregister(t *testing.T) {
	setupAssets(t)
	defer teardownAssets(t)

	s, err := New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatDocumentation,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Deregister()
	})
	assert.NoError(t, err, "deregister service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 7, "number of log lines mismatch")

	wanted := "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = `Running 'rm -f .+/traefik/etc/certs.d/rid:finchctl:.+\.pem' as .+@localhost`
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(context.Background(), Options{
		TargetURL:  "localhost",
		Format:     target.FormatJSON,
		DryRun:     true,
		CmdTimeout: 300 * time.Second,
	})
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Deregister()
	})
	assert.NoError(t, err, "deregister service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func capture(f func()) string {
	originalStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = originalStdout

	var buf = make([]byte, 10192)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func setupAssets(t *testing.T) {
	libDir, err := os.MkdirTemp("", "finch-test-lib-*")
	assert.NoError(t, err, "create temp lib dir")
	cfgDir, err := os.MkdirTemp("", "finchctl-test-cfg-*")
	assert.NoError(t, err, "create temp cfg dir")

	err = os.Setenv(ServiceLibEnv, libDir)
	assert.NoError(t, err, "set lib dir env")
	err = os.Setenv(config.ConfigLocationEnv, cfgDir)
	assert.NoError(t, err, "set cfg dir env")

	err = os.WriteFile(libDir+"/finch.json", []byte(`{ "hostname": "localhost" }`), 0600)
	assert.NoError(t, err, "write finch.json")
	err = os.WriteFile(cfgDir+"/finch.json", []byte(`{ "stacks": [ { "name": "localhost" } ] }`), 0600)
	assert.NoError(t, err, "write finch.json")
}

func teardownAssets(t *testing.T) {
	libDir := os.Getenv(ServiceLibEnv)
	cfgDir := os.Getenv(config.ConfigLocationEnv)

	err := os.RemoveAll(libDir)
	assert.NoError(t, err, "remove temp lib dir")
	err = os.RemoveAll(cfgDir)
	assert.NoError(t, err, "remove temp cfg dir")
}
