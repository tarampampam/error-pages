package config

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Templates []struct {
		Path    string `yaml:"path"`
		Name    string `yaml:"name"`
		Content string `yaml:"content"`
	} `yaml:"templates"`
	Pages map[string]struct {
		Message     string `yaml:"message"`
		Description string `yaml:"description"`
	} `yaml:"pages"`
}

// Validate the config and return an error if something is wrong.
func (c Config) Validate() error {
	if len(c.Templates) == 0 {
		return errors.New("empty templates list")
	} else {
		for i := 0; i < len(c.Templates); i++ {
			if c.Templates[i].Name == "" && c.Templates[i].Path == "" {
				return errors.New("empty path and name with index " + strconv.Itoa(i))
			}

			if c.Templates[i].Path == "" && c.Templates[i].Content == "" {
				return errors.New("empty path and template content with index " + strconv.Itoa(i))
			}
		}
	}

	if len(c.Pages) == 0 {
		return errors.New("empty pages list")
	} else {
		for code := range c.Pages {
			if code == "" {
				return errors.New("empty page code")
			}

			if strings.ContainsRune(code, ' ') {
				return errors.New("code should not contain whitespaces")
			}
		}
	}

	return nil
}

// LoadTemplates loading templates content from the local files and return it.
func (c Config) LoadTemplates() (map[string][]byte, error) {
	var templates = make(map[string][]byte) // map[template_name]template_content

	for i := 0; i < len(c.Templates); i++ {
		var name string

		if c.Templates[i].Name == "" {
			basename := filepath.Base(c.Templates[i].Path)
			name = strings.TrimSuffix(basename, filepath.Ext(basename))
		} else {
			name = c.Templates[i].Name
		}

		var content []byte

		if c.Templates[i].Content == "" {
			b, err := ioutil.ReadFile(c.Templates[i].Path)
			if err != nil {
				return nil, errors.Wrap(err, "cannot load content for the template "+name)
			}

			content = b
		} else {
			content = []byte(c.Templates[i].Content)
		}

		templates[name] = content
	}

	return templates, nil
}

// FromYaml creates new config instance using YAML-structured content.
func FromYaml(in []byte) (cfg *Config, err error) {
	cfg = &Config{}

	in, err = envsubst.Bytes(in)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(in, cfg); err != nil {
		return nil, errors.Wrap(err, "cannot parse configuration file")
	}

	return
}

// FromYamlFile creates new config instance using YAML file.
func FromYamlFile(filepath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read configuration file")
	}

	return FromYaml(bytes)
}
