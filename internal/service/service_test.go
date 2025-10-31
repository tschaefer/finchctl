/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package service

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tschaefer/finchctl/internal/target"
)

type track struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func Test_Deploy(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Deploy()
	})
	assert.NoError(t, err, "deploy service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 35, "number of log lines")

	wanted := "Running '[ \"${EUID:\\-$(id -u)}\" -eq 0 ] || command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'timeout 180 bash -c 'until curl -fs -o /dev/null -w \"%{http_code}\" http://localhost | grep -qE \"^[234][0-9]{2}$\"; do sleep 2; done'' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(nil, "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Deploy()
	})
	assert.NoError(t, err, "deploy service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running '[ \"${EUID:\\-$(id -u)}\" -eq 0 ] || command -v sudo' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Teardown(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Teardown()
	})
	assert.NoError(t, err, "teardown service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 3, "number of log lines mismatch")

	wanted := "Running 'sudo docker compose --file /var/lib/finch/docker-compose.yaml down --volumes' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo rm -rf /var/lib/finch' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(nil, "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "create service")

	record = capture(func() {
		err = s.Teardown()
	})
	assert.NoError(t, err, "teardown service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'sudo docker compose --file /var/lib/finch/docker-compose.yaml down --volumes' as .+@localhost"
	assert.Regexp(t, wanted, track.Message, "first log line")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Update(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create service")

	record := capture(func() {
		err = s.Update()
	})
	assert.NoError(t, err, "update service")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 28, "number of log lines")

	wanted := "Running 'sudo cat /var/lib/finch/finch.json' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo docker image prune --force' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	s, err = New(nil, "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "update service")

	record = capture(func() {
		err = s.Update()
	})
	assert.NoError(t, err, "update service")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	wanted = "Running 'sudo cat /var/lib/finch/finch.json' as .+@localhost"
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

	var buf = make([]byte, 5096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}
