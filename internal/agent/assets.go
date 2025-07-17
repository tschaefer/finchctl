/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package agent

import (
	"embed"
	"io/fs"
)

//go:embed assets/*
var assets embed.FS

var Assets fs.FS

func init() {
	var err error

	Assets, err = fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
}
