package templates

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed *.html
var content embed.FS

// BuiltIn returns a map of built-in templates. The key is the template name and the value is the template content.
func BuiltIn() map[string]string {
	var (
		list, _ = fs.ReadDir(content, ".") // error check is covered by unit tests
		result  = make(map[string]string, len(list))
	)

	for _, file := range list {
		if data, err := fs.ReadFile(content, file.Name()); err == nil {
			var (
				fileName     = filepath.Base(file.Name())
				ext          = filepath.Ext(fileName)
				templateName string
			)

			if ext != "" && fileName != ext {
				templateName = strings.TrimSuffix(fileName, ext)
			} else {
				templateName = fileName
			}

			result[templateName] = string(data)
		}
	}

	return result
}
