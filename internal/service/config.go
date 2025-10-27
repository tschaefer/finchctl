/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT license, see LICENSE in the project root for details.
*/
package service

type Config struct {
	CreatedAt   string `json:"created_at"`
	Id          string `json:"id"`
	Database    string `json:"database"`
	Secret      string `json:"secret"`
	Hostname    string `json:"hostname"`
	Version     string `json:"version"`
	Credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"credentials"`
}
