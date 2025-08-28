/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package agent

import (
	"os"
	"strings"
	"testing"

	"github.com/tschaefer/finchctl/internal/target"
)

func Test_Deploy(t *testing.T) {
	a, err := New("", "localhost", target.FormatDocumentation, true)
	if err != nil {
		t.Fatal(err)
	}

	record := capture(func() {
		err = a.Deploy()
	})
	if err != nil {
		t.Fatal(err)
	}

	tracks := strings.Split(record, "\n")
	if len(tracks) != 8 {
		t.Fatalf("expected at 8 lines of output, got %d", len(tracks))
	}

	wanted := "Running 'uname -sm' as tschaefer@localhost"
	got := tracks[0]
	if got != wanted {
		t.Fatalf("expected first line to be '%s', got '%s'", wanted, got)
	}

	wanted = "Running 'sudo systemctl enable --now alloy' as tschaefer@localhost"
	got = tracks[len(tracks)-2]
	if got != wanted {
		t.Fatalf("expected last line to be '%s', got '%s'", wanted, got)
	}
}

func Test_Teardown(t *testing.T) {
	a, err := New("", "localhost", target.FormatDocumentation, true)
	if err != nil {
		t.Fatal(err)
	}

	record := capture(func() {
		err = a.Teardown()
	})
	if err != nil {
		t.Fatal(err)
	}

	tracks := strings.Split(record, "\n")
	if len(tracks) != 6 {
		t.Fatalf("expected at 6 lines of output, got %d", len(tracks))
	}

	wanted := "Running 'sudo systemctl stop alloy.service' as tschaefer@localhost"
	got := tracks[0]
	if got != wanted {
		t.Fatalf("expected first line to be '%s', got '%s'", wanted, got)
	}

	wanted = "Running 'sudo rm -rf /var/lib/alloy' as tschaefer@localhost"
	got = tracks[len(tracks)-2]
	if got != wanted {
		t.Fatalf("expected last line to be '%s', got '%s'", wanted, got)
	}
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
