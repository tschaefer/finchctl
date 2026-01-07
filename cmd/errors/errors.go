/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tschaefer/finchctl/internal/target"
)

func CheckErr(msg any, format target.Format) {
	if msg == nil {
		return
	}

	switch format {
	case target.FormatJSON:
		data := map[string]string{
			"timestamp": time.Now().Format(time.RFC3339),
			"error":     fmt.Sprintf("%v", msg),
		}
		jsonData, _ := json.Marshal(data)
		fmt.Println(string(jsonData))
	default:
		fmt.Println(msg)
	}

	os.Exit(1)
}
