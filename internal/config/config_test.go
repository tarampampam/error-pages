package config_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/config"
)

func TestConfig_Validate(t *testing.T) {
	for name, tt := range map[string]struct {
		giveConfig func() config.Config
		wantError  error
	}{
		"valid": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{"foo", "bar", "baz"},
				}

				c.Pages = map[string]struct {
					Message     string `yaml:"message"`
					Description string `yaml:"description"`
				}{
					"400": {"Bad Request", "The server did not understand the request"},
				}

				return c
			},
			wantError: nil,
		},
		"empty templates list": {
			giveConfig: func() config.Config {
				return config.Config{}
			},
			wantError: errors.New("empty templates list"),
		},
		"empty path and name": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{
						Path:    "foo",
						Name:    "bar",
						Content: "baz",
					},
					{
						Path:    "",
						Name:    "",
						Content: "blah",
					},
				}

				return c
			},
			wantError: errors.New("empty path and name with index 1"),
		},
		"empty path and template content": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{
						Path:    "foo",
						Name:    "bar",
						Content: "baz",
					},
					{
						Path:    "",
						Name:    "blah",
						Content: "",
					},
				}

				return c
			},
			wantError: errors.New("empty path and template content with index 1"),
		},
		"empty pages list": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{"foo", "bar", "baz"},
				}

				c.Pages = map[string]struct {
					Message     string `yaml:"message"`
					Description string `yaml:"description"`
				}{}

				return c
			},
			wantError: errors.New("empty pages list"),
		},
		"empty page code": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{"foo", "bar", "baz"},
				}

				c.Pages = map[string]struct {
					Message     string `yaml:"message"`
					Description string `yaml:"description"`
				}{
					"": {"foo", "bar"},
				}

				return c
			},
			wantError: errors.New("empty page code"),
		},
		"code should not contain whitespaces": {
			giveConfig: func() config.Config {
				c := config.Config{}

				c.Templates = []struct {
					Path    string `yaml:"path"`
					Name    string `yaml:"name"`
					Content string `yaml:"content"`
				}{
					{"foo", "bar", "baz"},
				}

				c.Pages = map[string]struct {
					Message     string `yaml:"message"`
					Description string `yaml:"description"`
				}{
					" 123": {"foo", "bar"},
				}

				return c
			},
			wantError: errors.New("code should not contain whitespaces"),
		},
	} {
		tt := tt

		t.Run(name, func(t *testing.T) {
			err := tt.giveConfig().Validate()

			if tt.wantError != nil {
				assert.EqualError(t, err, tt.wantError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFromYaml(t *testing.T) {
	var cases = []struct { //nolint:maligned
		name          string
		giveYaml      []byte
		giveEnv       map[string]string
		wantErr       bool
		checkResultFn func(*testing.T, *config.Config)
	}{
		{
			name: "with all possible values",
			giveEnv: map[string]string{
				"__GHOST_PATH": "./templates/ghost.html",
				"__GHOST_NAME": "ghost",
			},
			giveYaml: []byte(`
templates:
  - path: ${__GHOST_PATH}
    name: ${__GHOST_NAME:-default_value} # name is optional
  - path: ./templates/l7-light.html
  - name: Foo
    content: |
      Some content
      New line

pages:
  400:
    message: Bad Request
    description: The server did not understand the request

  401:
    message: Unauthorized
    description: The requested page needs a username and a password
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, cfg *config.Config) {
				assert.Len(t, cfg.Templates, 3)
				assert.Equal(t, "./templates/ghost.html", cfg.Templates[0].Path)
				assert.Equal(t, "ghost", cfg.Templates[0].Name)
				assert.Equal(t, "", cfg.Templates[0].Content)
				assert.Equal(t, "./templates/l7-light.html", cfg.Templates[1].Path)
				assert.Equal(t, "", cfg.Templates[1].Name)
				assert.Equal(t, "", cfg.Templates[1].Content)
				assert.Equal(t, "", cfg.Templates[2].Path)
				assert.Equal(t, "Foo", cfg.Templates[2].Name)
				assert.Equal(t, "Some content\nNew line\n", cfg.Templates[2].Content)

				assert.Len(t, cfg.Pages, 2)
				assert.Equal(t, "Bad Request", cfg.Pages["400"].Message)
				assert.Equal(t, "The server did not understand the request", cfg.Pages["400"].Description)
				assert.Equal(t, "Unauthorized", cfg.Pages["401"].Message)
				assert.Equal(t, "The requested page needs a username and a password", cfg.Pages["401"].Description)
			},
		},
		{
			name:     "broken yaml",
			giveYaml: []byte(`foo bar`),
			wantErr:  true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.giveEnv != nil {
				for key, value := range tt.giveEnv {
					assert.NoError(t, os.Setenv(key, value))
				}
			}

			conf, err := config.FromYaml(tt.giveYaml)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				tt.checkResultFn(t, conf)
			}

			if tt.giveEnv != nil {
				for key := range tt.giveEnv {
					assert.NoError(t, os.Unsetenv(key))
				}
			}
		})
	}
}

func TestFromYamlFile(t *testing.T) {
	var cases = []struct { //nolint:maligned
		name             string
		giveYamlFilePath string
		wantErr          bool
		checkResultFn    func(*testing.T, *config.Config)
	}{
		{
			name:             "with all possible values",
			giveYamlFilePath: "./testdata/simple.yml",
			wantErr:          false,
			checkResultFn: func(t *testing.T, cfg *config.Config) {
				assert.Len(t, cfg.Templates, 2)
				assert.Equal(t, "./templates/ghost.html", cfg.Templates[0].Path)
				assert.Equal(t, "ghost", cfg.Templates[0].Name)
				assert.Equal(t, "./templates/l7-light.html", cfg.Templates[1].Path)
				assert.Equal(t, "", cfg.Templates[1].Name)

				assert.Len(t, cfg.Pages, 2)
				assert.Equal(t, "Bad Request", cfg.Pages["400"].Message)
				assert.Equal(t, "The server did not understand the request", cfg.Pages["400"].Description)
				assert.Equal(t, "Unauthorized", cfg.Pages["401"].Message)
				assert.Equal(t, "The requested page needs a username and a password", cfg.Pages["401"].Description)
			},
		},
		{
			name:             "broken yaml",
			giveYamlFilePath: "./testdata/broken.yml",
			wantErr:          true,
		},
		{
			name:             "wrong file path",
			giveYamlFilePath: "foo bar",
			wantErr:          true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conf, err := config.FromYamlFile(tt.giveYamlFilePath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				tt.checkResultFn(t, conf)
			}
		})
	}
}
