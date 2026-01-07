/*
Copyright (c) Tobias Schäfer. All rights reserved.
Licensed under the MIT License, see LICENSE file in the project root for details.
*/
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func raiseError() error {
	return &DeployServiceError{Message: "deployment failed", Reason: "insufficient resources"}
}

func Test_convertError(t *testing.T) {
	err := raiseError()
	assert.Error(t, err, "expected an error from raiseError function")
	assert.IsType(t, &DeployServiceError{}, err, "expected error to be of type DeployServiceError")

	err = convertError(err, &UpdateServiceError{})
	assert.Error(t, err, "expected an error after conversion")
	assert.IsType(t, &UpdateServiceError{}, err, "expected error to be of type UpdateServiceError")
	assert.Equal(t, "deployment failed", err.(*UpdateServiceError).Message, "expected message to be preserved")
	assert.Equal(t, "insufficient resources", err.(*UpdateServiceError).Reason, "expected reason to be preserved")

	err = convertError(nil, &UpdateServiceError{})
	assert.Nil(t, err, "expected nil when input error is nil")

	err = convertError(err, nil)
	assert.Nil(t, err, "expected nil when target error type is nil")
}
