/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package version

import (
	"fmt"
	"os"

	"github.com/denisbrodbeck/machineid"
)

var (
	GitCommit, Version string
)

func Release() string {
	if Version == "" {
		Version = "dev"
	}

	return Version
}

func Commit() string {
	return GitCommit
}

func ResourceId() string {
	id, err := machineid.ProtectedID("finchctl")
	if err != nil {
		panic(fmt.Errorf("failed to read machine ID: %w", err))
	}

	return fmt.Sprintf("rid:finchctl:%s", id[0:16])
}

func Banner() string {
	return `
  __ _            _          _   _ 
 / _(_)_ __   ___| |__   ___| |_| |
| |_| | '_ \ / __| '_ \ / __| __| |
|  _| | | | | (__| | | | (__| |_| |
|_| |_|_| |_|\___|_| |_|\___|\__|_|
`
}

func Print() {
	no_color := os.Getenv("NO_COLOR")
	if no_color != "" {
		fmt.Printf("%s\n", Banner())
	} else {
		fmt.Printf("\033[34m%s\033[0m\n", Banner())
	}
	fmt.Printf("Release:    %s\n", Release())
	fmt.Printf("Commit:     %s\n", Commit())
	fmt.Printf("ResourceID: %s\n", ResourceId())
}
