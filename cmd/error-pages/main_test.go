package main

import (
	"os"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
)

func Test_Main(t *testing.T) {
	os.Args = []string{"", "--help"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	output := capturer.CaptureStdout(main)

	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "Flags:")
}

func Test_MainWithoutCommands(t *testing.T) {
	os.Args = []string{""}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	output := capturer.CaptureStdout(main)

	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "Flags:")
}

func Test_MainUnknownSubcommand(t *testing.T) {
	os.Args = []string{"", "foobar"}
	exitFn = func(code int) { assert.Equal(t, 1, code) }

	output := capturer.CaptureStderr(main)

	assert.Contains(t, output, "unknown command")
	assert.Contains(t, output, "foobar")
}
