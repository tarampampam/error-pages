package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type templates map[string]string // map[name]content

// Add adds a new template.
func (tpl templates) Add(name, content string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	tpl[name] = content

	return nil
}

// AddFromFile reads the file content and adds it as a new template.
func (tpl templates) AddFromFile(path string, name ...string) (addedTemplateName string, _ error) {
	// check if the file exists and is not a directory
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file %s not found", path)
		}

		return "", err
	} else if stat.IsDir() {
		return "", fmt.Errorf("%s is not a file", path)
	}

	// read the file content
	var content, err = os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s: %w", path, err)
	}

	var templateName string

	if len(name) > 0 && name[0] != "" { // if the name is provided, use it
		templateName = name[0]
	} else { // otherwise, use the file name without the extension
		var (
			fileName = filepath.Base(path)
			ext      = filepath.Ext(fileName)
		)

		if ext != "" && fileName != ext {
			templateName = strings.TrimSuffix(fileName, ext)
		} else {
			templateName = fileName
		}
	}

	// add the template to the config
	tpl[templateName] = string(content)

	return templateName, nil
}

// Names returns all template names sorted alphabetically.
func (tpl templates) Names() []string {
	var names = make([]string, 0, len(tpl))

	for name := range tpl {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// Has checks if the template with the specified name exists.
func (tpl templates) Has(name string) (found bool) { _, found = tpl[name]; return } //nolint:nlreturn

// Get returns the template content by the specified name, if it exists.
func (tpl templates) Get(name string) (data string, ok bool) { data, ok = tpl[name]; return } //nolint:nlreturn
