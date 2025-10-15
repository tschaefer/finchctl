/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package format

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tschaefer/finchctl/internal/target"
)

func GetRunFormat(cmd *cobra.Command) (target.Format, error) {
	formatter, _ := cmd.Flags().GetString("run.format")

	var format target.Format
	var err error
	switch formatter {
	case "documentation":
		format = target.FormatDocumentation
	case "quiet":
		format = target.FormatQuiet
	case "progress":
		format = target.FormatProgress
	default:
		err = fmt.Errorf("unknown format %s", formatter)
	}

	return format, err
}
