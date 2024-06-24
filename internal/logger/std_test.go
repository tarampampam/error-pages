package logger_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/logger"
)

func TestNewStdLog(t *testing.T) {
	var (
		buf    bytes.Buffer
		log, _ = logger.New(logger.InfoLevel, logger.JSONFormat, &buf)

		std = logger.NewStdLog(log)
	)

	std.Print("test")

	assert.Contains(t, buf.String(), "test")
}
