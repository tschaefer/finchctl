/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package target

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
)

type Format int64

const (
	FormatQuiet         Format = 0
	FormatProgress      Format = 1
	FormatDocumentation Format = 2
)

type Target interface {
	Run(command string) ([]byte, error)
	Copy(src, dest, mode, owner string) error
	Request(method string, url *url.URL, data []byte) ([]byte, error)
}

func NewTarget(hostUrl string, format Format, dryRun bool) (Target, error) {
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
		return NewLocal(host, format, dryRun)
	}

	return NewRemote(host, format, dryRun)
}

func parseHostUrl(hostUrl string) (host *url.URL, err error) {
	if !strings.HasPrefix(hostUrl, "host://") {
		hostUrl = "host://" + hostUrl
	}
	host, err = url.Parse(hostUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid host URL: %w", err)
	}

	user := host.User.Username()
	if user == "" {
		user = os.Getenv("USER")
	}
	host.User = url.User(user)
	port := host.Port()
	if port == "" {
		host.Host = fmt.Sprintf("%s:%s", host.Hostname(), "22")
	}

	return host, nil
}

func printProgress(message string, format Format) {
	switch format {
	case FormatProgress:
		fmt.Print(".")
	case FormatDocumentation:
		fmt.Println(message)
	case FormatQuiet:
		// Do nothing
	default:
		fmt.Println(".")
	}
}
