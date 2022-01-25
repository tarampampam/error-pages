package config

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is a main (exportable) config struct.
type Config struct {
	Templates []Template
	Pages     map[string]Page   // map key is a page code
	Formats   map[string]Format // map key is a format name
}

// Template returns a Template with the passes name.
func (c *Config) Template(name string) (*Template, bool) {
	for i := 0; i < len(c.Templates); i++ {
		if c.Templates[i].name == name {
			return &c.Templates[i], true
		}
	}

	return &Template{}, false
}

func (c *Config) JSONFormat() (*Format, bool) { return c.format("json") }
func (c *Config) XMLFormat() (*Format, bool)  { return c.format("xml") }

func (c *Config) format(name string) (*Format, bool) {
	if f, ok := c.Formats[name]; ok {
		if len(f.content) > 0 {
			return &f, true
		}
	}

	return &Format{}, false
}

// TemplateNames returns all template names.
func (c *Config) TemplateNames() []string {
	n := make([]string, len(c.Templates))

	for i, t := range c.Templates {
		n[i] = t.name
	}

	return n
}

// Template describes HTTP error page template.
type Template struct {
	name    string
	content []byte
}

// Name returns the name of the template.
func (t Template) Name() string { return t.name }

// Content returns the template content.
func (t Template) Content() []byte { return t.content }

func (t *Template) loadContentFromFile(filePath string) (err error) {
	if t.content, err = ioutil.ReadFile(filePath); err != nil {
		return errors.Wrap(err, "cannot load content for the template "+t.Name()+" from file "+filePath)
	}

	return
}

// Page describes error page.
type Page struct {
	code        string
	message     string
	description string
}

// Code returns the code of the Page.
func (p Page) Code() string { return p.code }

// Message returns the message of the Page.
func (p Page) Message() string { return p.message }

// Description returns the description of the Page.
func (p Page) Description() string { return p.description }

// Format describes different response formats.
type Format struct {
	name    string
	content []byte
}

// Name returns the name of the format.
func (f Format) Name() string { return f.name }

// Content returns the format content.
func (f Format) Content() []byte { return f.content }

// config is internal struct for marshaling/unmarshaling configuration file content.
type config struct {
	Templates []struct {
		Path    string `yaml:"path"`
		Name    string `yaml:"name"`
		Content string `yaml:"content"`
	} `yaml:"templates"`

	Formats map[string]struct {
		Content string `yaml:"content"`
	} `yaml:"formats"`

	Pages map[string]struct {
		Message     string `yaml:"message"`
		Description string `yaml:"description"`
	} `yaml:"pages"`
}

// Validate the config struct and return an error if something is wrong.
func (c config) Validate() error {
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

	if len(c.Formats) > 0 {
		for name := range c.Formats {
			if name == "" {
				return errors.New("empty format name")
			}

			if strings.ContainsRune(name, ' ') {
				return errors.New("format should not contain whitespaces")
			}
		}
	}

	return nil
}

// Export the config struct into Config.
func (c *config) Export() (*Config, error) {
	cfg := &Config{}

	cfg.Templates = make([]Template, 0, len(c.Templates))

	for i := 0; i < len(c.Templates); i++ {
		tpl := Template{name: c.Templates[i].Name}

		if c.Templates[i].Content == "" {
			if c.Templates[i].Path == "" {
				return nil, errors.New("path to the template " + c.Templates[i].Name + " not provided")
			}

			if err := tpl.loadContentFromFile(c.Templates[i].Path); err != nil {
				return nil, err
			}
		} else {
			tpl.content = []byte(c.Templates[i].Content)
		}

		cfg.Templates = append(cfg.Templates, tpl)
	}

	cfg.Pages = make(map[string]Page, len(c.Pages))

	for code, p := range c.Pages {
		cfg.Pages[code] = Page{code: code, message: p.Message, description: p.Description}
	}

	cfg.Formats = make(map[string]Format, len(c.Formats))

	for name, f := range c.Formats {
		cfg.Formats[name] = Format{name: name, content: []byte(strings.TrimSpace(f.Content))}
	}

	return cfg, nil
}

// FromYaml creates new Config instance using YAML-structured content.
func FromYaml(in []byte) (_ *Config, err error) {
	in, err = envsubst.Bytes(in)
	if err != nil {
		return nil, err
	}

	c := &config{}

	if err = yaml.Unmarshal(in, c); err != nil {
		return nil, errors.Wrap(err, "cannot parse configuration file")
	}

	var basename string

	for i := 0; i < len(c.Templates); i++ {
		if c.Templates[i].Name == "" { // set the template name from file path
			basename = filepath.Base(c.Templates[i].Path)
			c.Templates[i].Name = strings.TrimSuffix(basename, filepath.Ext(basename))
		}
	}

	if err = c.Validate(); err != nil {
		return nil, err
	}

	return c.Export()
}

// FromYamlFile creates new Config instance using YAML file.
func FromYamlFile(filepath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read configuration file")
	}

	// the following code makes it possible to use the relative links in the config file (`.` means "directory with
	// the config file")
	cwd, err := os.Getwd()
	if err == nil {
		if err = os.Chdir(path.Dir(filepath)); err != nil {
			return nil, err
		}

		defer func() { _ = os.Chdir(cwd) }()
	}

	return FromYaml(bytes)
}
