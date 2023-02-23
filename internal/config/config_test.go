package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/config"
)

func TestFromYaml(t *testing.T) {
	var cases = map[string]struct { //nolint:maligned
		giveYaml      []byte
		giveEnv       map[string]string
		wantErr       bool
		checkResultFn func(*testing.T, *config.Config)
	}{
		"with all possible values": {
			giveEnv: map[string]string{
				"__FOO_TPL_PATH": "./testdata/foo-tpl.html",
				"__FOO_TPL_NAME": "Foo Template",
			},
			giveYaml: []byte(`
templates:
  - path: ${__FOO_TPL_PATH}
    name: ${__FOO_TPL_NAME:-default_value} # name is optional
  - path: ./testdata/bar-tpl.html
  - name: Baz
    content: |
      Some content {{ code }}
      New line

formats:
  json:
    content: |
      {"code": "{{code}}"}
  Avada_Kedavra:
    content: "{{ message }}"

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

				tpl, found := cfg.Template("Foo Template")
				assert.True(t, found)
				assert.Equal(t, "Foo Template", tpl.Name())
				assert.Equal(t, "<html><body>foo {{ code }}</body></html>\n", string(tpl.Content()))

				tpl, found = cfg.Template("bar-tpl")
				assert.True(t, found)
				assert.Equal(t, "bar-tpl", tpl.Name())
				assert.Equal(t, "<html><body>bar {{ code }}</body></html>\n", string(tpl.Content()))

				tpl, found = cfg.Template("Baz")
				assert.True(t, found)
				assert.Equal(t, "Baz", tpl.Name())
				assert.Equal(t, "Some content {{ code }}\nNew line\n", string(tpl.Content()))

				tpl, found = cfg.Template("NonExists")
				assert.False(t, found)
				assert.Equal(t, "", tpl.Name())
				assert.Equal(t, "", string(tpl.Content()))

				assert.Len(t, cfg.Formats, 2)

				format, found := cfg.Formats["json"]
				assert.True(t, found)
				assert.Equal(t, `{"code": "{{code}}"}`, string(format.Content()))

				format, found = cfg.Formats["Avada_Kedavra"]
				assert.True(t, found)
				assert.Equal(t, "{{ message }}", string(format.Content()))

				assert.Len(t, cfg.Pages, 2)

				errPage, found := cfg.Pages["400"]
				assert.True(t, found)
				assert.Equal(t, "400", errPage.Code())
				assert.Equal(t, "Bad Request", errPage.Message())
				assert.Equal(t, "The server did not understand the request", errPage.Description())

				errPage, found = cfg.Pages["401"]
				assert.True(t, found)
				assert.Equal(t, "401", errPage.Code())
				assert.Equal(t, "Unauthorized", errPage.Message())
				assert.Equal(t, "The requested page needs a username and a password", errPage.Description())

				errPage, found = cfg.Pages["666"]
				assert.False(t, found)
				assert.Equal(t, "", errPage.Message())
				assert.Equal(t, "", errPage.Code())
				assert.Equal(t, "", errPage.Description())
			},
		},
		"broken yaml": {
			giveYaml: []byte(`foo bar`),
			wantErr:  true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
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
	var cases = map[string]struct { //nolint:maligned
		giveYamlFilePath string
		wantErr          bool
		checkResultFn    func(*testing.T, *config.Config)
	}{
		"with all possible values": {
			giveYamlFilePath: "./testdata/simple.yml",
			wantErr:          false,
			checkResultFn: func(t *testing.T, cfg *config.Config) {
				assert.Len(t, cfg.Templates, 2)

				tpl, found := cfg.Template("ghost")
				assert.True(t, found)
				assert.Equal(t, "ghost", tpl.Name())
				assert.Equal(t, "<html><body>foo {{ code }}</body></html>\n", string(tpl.Content()))

				tpl, found = cfg.Template("bar-tpl")
				assert.True(t, found)
				assert.Equal(t, "bar-tpl", tpl.Name())
				assert.Equal(t, "<html><body>bar {{ code }}</body></html>\n", string(tpl.Content()))

				assert.Len(t, cfg.Pages, 2)

				errPage, found := cfg.Pages["400"]
				assert.True(t, found)
				assert.Equal(t, "400", errPage.Code())
				assert.Equal(t, "Bad Request", errPage.Message())
				assert.Equal(t, "The server did not understand the request", errPage.Description())

				errPage, found = cfg.Pages["401"]
				assert.True(t, found)
				assert.Equal(t, "401", errPage.Code())
				assert.Equal(t, "Unauthorized", errPage.Message())
				assert.Equal(t, "The requested page needs a username and a password", errPage.Description())
			},
		},
		"broken yaml": {
			giveYamlFilePath: "./testdata/broken.yml",
			wantErr:          true,
		},
		"wrong file path": {
			giveYamlFilePath: "foo bar",
			wantErr:          true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
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
