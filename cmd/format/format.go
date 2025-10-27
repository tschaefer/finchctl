/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package format

import (
	"fmt"

	"github.com/tschaefer/finchctl/internal/target"
)

func GetRunFormat(name string) (target.Format, error) {
	var format target.Format
	var err error
	switch name {
	case "documentation":
		format = target.FormatDocumentation
	case "quiet":
		format = target.FormatQuiet
	case "progress":
		format = target.FormatProgress
	default:
		err = fmt.Errorf("unknown format %s", name)
	}

	return format, err
}
