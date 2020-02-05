package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArguments(t *testing.T) {
	var stdout strings.Builder

	const test = "Hello, World!"

	err := New("echo").
		WithArguments("-n", test).
		WithStdout(&stdout).
		Run()
	assert.NoError(t, err)

	assert.Equal(t, test, stdout.String())
}
