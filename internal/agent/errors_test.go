/*
Copyright (c) 2025 Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func raiseError() error {
	return &DeployAgentError{Message: "deployment failed", Reason: "insufficient resources"}
}

func Test_convertError(t *testing.T) {
	err := raiseError()
	assert.Error(t, err, "expected an error from raiseError function")
	assert.IsType(t, &DeployAgentError{}, err, "expected error to be of type DeployAgentError")

	err = convertError(err, &UpdateAgentError{})
	assert.Error(t, err, "expected an error after conversion")
	assert.IsType(t, &UpdateAgentError{}, err, "expected error to be of type UpdateAgentError")
	assert.Equal(t, "deployment failed", err.(*UpdateAgentError).Message, "expected message to be preserved")
	assert.Equal(t, "insufficient resources", err.(*UpdateAgentError).Reason, "expected reason to be preserved")

	err = convertError(nil, &UpdateAgentError{})
	assert.Nil(t, err, "expected nil when input error is nil")

	err = convertError(err, nil)
	assert.Nil(t, err, "expected nil when target error type is nil")
}
