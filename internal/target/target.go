/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"slices"
	"strings"
	"time"
)

type Format int64

const (
	FormatQuiet         Format = 0
	FormatProgress      Format = 1
	FormatDocumentation Format = 2
	FormatJSON          Format = 3
)

type Target interface {
	Run(ctx context.Context, command string) ([]byte, error)
	Copy(ctx context.Context, src, dest, mode, owner string) error
}

type Options struct {
	Format     Format
	DryRun     bool
	CmdTimeout time.Duration
}

func New(hostUrl string, opts Options) (Target, error) {
	host, err := parseHostUrl(hostUrl)
	if err != nil {
		return nil, err
	}

	local := []string{
		"localhost",
		"local",
		"127.0.0.1",
		"::1",
	}
	if slices.Contains(local, host.Hostname()) {
		return newLocal(host, opts)
	}

	return newRemote(host, opts)
}

func parseHostUrl(hostUrl string) (host *url.URL, err error) {
	if !strings.HasPrefix(hostUrl, "host://") {
		hostUrl = "host://" + hostUrl
	}
	host, err = url.Parse(hostUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %w", err)
	}

	username := host.User.Username()
	if username == "" {
		username = "unknown"
		user, err := user.Current()
		if err == nil {
			username = user.Username
		}
	}
	host.User = url.User(username)
	port := host.Port()
	if port == "" {
		host.Host = fmt.Sprintf("%s:%s", host.Hostname(), "22")
	}

	return host, nil
}

func PrintProgress(message string, format Format) {
	switch format {
	case FormatProgress:
		fmt.Print(".")
	case FormatDocumentation:
		fmt.Println(message)
	case FormatJSON:
		data := map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
			"message":   message,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	case FormatQuiet:
		// Do nothing
	default:
		fmt.Println(".")
	}
}
