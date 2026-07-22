/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func capture(f func(string, Format), m string, o Format) string {
	originalStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	f(m, o)

	_ = w.Close()
	os.Stdout = originalStdout

	var buf = make([]byte, 5096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

func Test_NewReturnsErrorIfHostnameIsInvalid(t *testing.T) {
	_, err := New("user@host:service", Options{})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid host URL", "error message should contain invalid URL")
}

func Test_NewReturnsLocalTargetIfHostIsLocal(t *testing.T) {
	hosts := []string{
		"localhost",
		"local",
		"127.0.0.1",
		"[::1]",
	}
	for _, host := range hosts {
		target, err := New(host, Options{})

		assert.NoError(t, err)
		assert.IsType(t, &local{}, target, "target should be type local")
	}
}

func Test_NewReturnsRemoteTargetIfHostIsNotLocal(t *testing.T) {
	orig := newRemoteTarget
	defer func() { newRemoteTarget = orig }()

	called := false
	newRemoteTarget = func(host *url.URL, opts Options) (Target, error) {
		called = true
		return &remote{Host: host.Hostname(), User: host.User.Username()}, nil
	}

	target, err := New("example.com", Options{})

	assert.NoError(t, err)
	assert.True(t, called, "newRemote should be called for non-local hosts")
	assert.IsType(t, &remote{}, target, "target should be type remote")
}

func Test_PrintProgressPrintsMessageOfRequestedFormat(t *testing.T) {
	message := "Hello, World."
	now := time.Now().Format(time.RFC3339)
	formats := map[Format]string{
		FormatProgress:      ".",
		FormatQuiet:         "",
		FormatJSON:          "{\"message\":\"" + message + "\",\"timestamp\":\"" + now + "\"}\n",
		FormatDocumentation: message + "\n",
	}
	for format, expected := range formats {
		actual := capture(PrintProgress, message, format)
		assert.Equal(t, actual, expected, "output should be formated")
	}
}
