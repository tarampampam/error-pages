package logger_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tarampampam/error-pages/internal/logger"
)

func TestFormat_String(t *testing.T) {
	for name, tt := range map[string]struct {
		giveFormat logger.Format
		wantString string
	}{
		"json":      {giveFormat: logger.JSONFormat, wantString: "json"},
		"console":   {giveFormat: logger.ConsoleFormat, wantString: "console"},
		"<unknown>": {giveFormat: logger.Format(255), wantString: "format(255)"},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.wantString, tt.giveFormat.String())
		})
	}
}

func TestParseFormat(t *testing.T) {
	for name, tt := range map[string]struct {
		giveBytes  []byte
		giveString string
		wantFormat logger.Format
		wantError  error
	}{
		"<empty value>":          {giveBytes: []byte(""), wantFormat: logger.ConsoleFormat},
		"<empty value> (string)": {giveString: "", wantFormat: logger.ConsoleFormat},
		"console":                {giveBytes: []byte("console"), wantFormat: logger.ConsoleFormat},
		"console (string)":       {giveString: "console", wantFormat: logger.ConsoleFormat},
		"json":                   {giveBytes: []byte("json"), wantFormat: logger.JSONFormat},
		"json (string)":          {giveString: "json", wantFormat: logger.JSONFormat},
		"foobar":                 {giveBytes: []byte("foobar"), wantError: errors.New("unrecognized logging format: \"foobar\"")}, //nolint:lll
	} {
		t.Run(name, func(t *testing.T) {
			var (
				f   logger.Format
				err error
			)

			if tt.giveString != "" {
				f, err = logger.ParseFormat(tt.giveString)
			} else {
				f, err = logger.ParseFormat(tt.giveBytes)
			}

			if tt.wantError == nil {
				require.NoError(t, err)
				require.Equal(t, tt.wantFormat, f)
			} else {
				require.EqualError(t, err, tt.wantError.Error())
			}
		})
	}
}
