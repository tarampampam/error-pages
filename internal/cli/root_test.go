package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/cli"
)

func TestSubcommands(t *testing.T) {
	cmd := cli.NewCommand("unit test")

	cases := []struct {
		giveName string
	}{
		{giveName: "build"},
		{giveName: "version"},
		{giveName: "healthcheck"},
		{giveName: "serve"},
	}

	// get all existing subcommands and put into the map
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = sub
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			if _, exists := subcommands[tt.giveName]; !exists {
				assert.Failf(t, "command not found", "command [%s] was not found", tt.giveName)
			}
		})
	}
}

func TestFlags(t *testing.T) {
	cmd := cli.NewCommand("unit test")

	cases := []struct {
		giveName      string
		wantShorthand string
		wantDefault   string
	}{
		{giveName: "verbose", wantShorthand: "v", wantDefault: "false"},
		{giveName: "debug", wantShorthand: "", wantDefault: "false"},
		{giveName: "log-json", wantShorthand: "", wantDefault: "false"},
		{giveName: "config-file", wantShorthand: "c", wantDefault: "./error-pages.yml"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag [%s] was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
			assert.Equal(t, tt.wantDefault, flag.DefValue)
		})
	}
}

func TestExecuting(t *testing.T) {
	cmd := cli.NewCommand("unit test")
	cmd.SetArgs([]string{})

	var executed bool

	if cmd.Run == nil { // override "Run" property for test (if it was not set)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			executed = true
		}
	}

	assert.NoError(t, cmd.Execute())
	assert.True(t, executed)
}
