package build

import (
	"bytes"
	_ "embed"
	"os"
	"sort"
	"text/template"
)

type (
	buildingHistory struct {
		items map[string][]historyItem
	}

	historyItem struct {
		Code, Message, Path string
	}
)

func newBuildingHistory() buildingHistory {
	return buildingHistory{items: make(map[string][]historyItem)}
}

func (bh *buildingHistory) Append(templateName, pageCode, message, path string) {
	if _, ok := bh.items[templateName]; !ok {
		bh.items[templateName] = make([]historyItem, 0)
	}

	bh.items[templateName] = append(bh.items[templateName], historyItem{
		Code:    pageCode,
		Message: message,
		Path:    path,
	})

	sort.Slice(bh.items[templateName], func(i, j int) bool { // keep history items sorted
		return bh.items[templateName][i].Code < bh.items[templateName][j].Code
	})
}

//go:embed index.tpl.html
var indexPageTemplate string //nolint:gochecknoglobals

func (bh *buildingHistory) WriteIndexFile(path string, perm os.FileMode) error {
	t, err := template.New("index").Parse(indexPageTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err = t.Execute(&buf, bh.items); err != nil {
		return err
	}

	defer buf.Reset() // optimization (is needed here?)

	return os.WriteFile(path, buf.Bytes(), perm)
}
