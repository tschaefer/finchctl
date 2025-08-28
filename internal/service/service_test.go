/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package service

import (
	"os"
	"strings"
	"testing"

	"github.com/tschaefer/finchctl/internal/target"
)

func Test_Deploy(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	if err != nil {
		t.Fatal(err)
	}

	record := capture(func() {
		err = s.Deploy()
	})
	if err != nil {
		t.Fatal(err)
	}

	tracks := strings.Split(record, "\n")
	if len(tracks) != 28 {
		t.Fatalf("expected at 28 lines of output, got %d", len(tracks))
	}

	wanted := "Running '[ \"${EUID:-$(id -u)}\" -eq 0 ] || command -v sudo' as tschaefer@localhost"
	got := tracks[0]
	if got != wanted {
		t.Fatalf("expected first line to be '%s', got '%s'", wanted, got)
	}

	wanted = "Running 'timeout 180 bash -c 'until curl -fs -o /dev/null -w \"%{http_code}\" http://localhost | grep -qE \"^[234][0-9]{2}$\"; do sleep 2; done'' as tschaefer@localhost"
	got = tracks[len(tracks)-2]
	if got != wanted {
		t.Fatalf("expected last line to be '%s', got '%s'", wanted, got)
	}
}

func Test_Teardown(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	if err != nil {
		t.Fatal(err)
	}

	record := capture(func() {
		err = s.Teardown()
	})
	if err != nil {
		t.Fatal(err)
	}

	tracks := strings.Split(record, "\n")
	if len(tracks) != 3 {
		t.Fatalf("expected at 3 lines of output, got %d", len(tracks))
	}

	wanted := "Running 'sudo docker compose --file /var/lib/finch/docker-compose.yml down --volumes' as tschaefer@localhost"
	got := tracks[0]
	if got != wanted {
		t.Fatalf("expected first line to be '%s', got '%s'", wanted, got)
	}

	wanted = "Running 'sudo rm -rf /var/lib/finch' as tschaefer@localhost"
	got = tracks[len(tracks)-2]
	if got != wanted {
		t.Fatalf("expected last line to be '%s', got '%s'", wanted, got)
	}
}

func Test_Update(t *testing.T) {
	s, err := New(nil, "localhost", target.FormatDocumentation, true)
	if err != nil {
		t.Fatal(err)
	}

	record := capture(func() {
		err = s.Update()
	})
	if err != nil {
		t.Fatal(err)
	}

	tracks := strings.Split(record, "\n")
	if len(tracks) != 7 {
		t.Fatalf("expected at 7 lines of output, got %d", len(tracks))
	}

	wanted := "Running 'sudo cat /var/lib/finch/finch.json' as tschaefer@localhost"
	got := tracks[0]
	if got != wanted {
		t.Fatalf("expected first line to be '%s', got '%s'", wanted, got)
	}

	wanted = "Running 'timeout 180 bash -c 'until curl -fs -o /dev/null -w \"%{http_code}\" http://localhost | grep -qE \"^[234][0-9]{2}$\"; do sleep 2; done'' as tschaefer@localhost"
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
