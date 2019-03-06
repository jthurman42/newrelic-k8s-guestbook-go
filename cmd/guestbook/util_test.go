package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogIfError(t *testing.T) {
	// No error returns nil
	assert.Nil(t, logIfError(nil, nil))

	// No error, with data, returns data
	assert.Equal(t, 1, logIfError(1, nil))

	// Error, no data, returns nil
	assert.Nil(t, logIfError(nil, errors.New("Test Error")))

	// Error, data, returns data
	assert.Equal(t, "asdf", logIfError("asdf", errors.New("Test Error with data")))
}
