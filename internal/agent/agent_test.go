/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package agent

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
	a, err := New("", "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create agent")

	record := capture(func() {
		err = a.Deploy()
	})
	assert.NoError(t, err)

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 10, "number of log lines")

	wanted := "Running 'uname -sm' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo systemctl enable --now alloy' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	a, err = New("", "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "create agent")

	record = capture(func() {
		err = a.Deploy()
	})
	assert.NoError(t, err)

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	assert.Regexp(t, "Running 'uname -sm' as .+@localhost", track.Message, "first log line message")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Teardown(t *testing.T) {
	a, err := New("", "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create agent")

	record := capture(func() {
		err = a.Teardown()
	})
	assert.NoError(t, err, "teardown agent")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 6, "number of log lines mismatch")

	wanted := "Running 'sudo systemctl stop alloy.service' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo rm -rf /var/lib/alloy' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	a, err = New("", "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "create agent")

	record = capture(func() {
		err = a.Teardown()
	})
	assert.NoError(t, err, "teardown agent")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	assert.Regexp(t, "Running 'sudo systemctl stop alloy.service' as .+@localhost", track.Message, "first log line message")
	assert.NotEmpty(t, track.Timestamp, "first log line timestamp")
}

func Test_Update(t *testing.T) {
	a, err := New("finch-agent.conf", "localhost", target.FormatDocumentation, true)
	assert.NoError(t, err, "create agent")

	record := capture(func() {
		err = a.Update()
	})
	assert.NoError(t, err, "update agent")

	tracks := strings.Split(record, "\n")
	assert.Len(t, tracks, 3, "number of log lines mismatch")

	wanted := "Copying from 'finch-agent.conf' to '/etc/alloy/alloy.config' as .+@localhost"
	assert.Regexp(t, wanted, tracks[0], "first log line")

	wanted = "Running 'sudo systemctl restart alloy.service' as .+@localhost"
	assert.Regexp(t, wanted, tracks[len(tracks)-2], "last log line")

	a, err = New("finch-agent.conf", "localhost", target.FormatJSON, true)
	assert.NoError(t, err, "create agent")

	record = capture(func() {
		err = a.Update()
	})
	assert.NoError(t, err, "update agent")

	tracks = strings.Split(record, "\n")
	var track track
	err = json.Unmarshal([]byte(tracks[0]), &track)
	assert.NoError(t, err, "unmarshal json output")

	assert.Regexp(t, "Copying from 'finch-agent.conf' to '/etc/alloy/alloy.config' as .+@localhost", track.Message, "first log line message")
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
