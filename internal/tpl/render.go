package tpl

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/tarampampam/error-pages/internal/version"
)

// These functions are always allowed in the templates.
var tplFnMap = template.FuncMap{ //nolint:gochecknoglobals
	"now":      time.Now,
	"hostname": os.Hostname,
	"json":     func(v interface{}) string { b, _ := json.Marshal(v); return string(b) }, //nolint:nlreturn
	"version":  version.Version,
	"int": func(v interface{}) int {
		if s, ok := v.(string); ok {
			if i, err := strconv.Atoi(s); err == nil {
				return i
			}
		} else if i, ok := v.(int); ok {
			return i
		}

		return 0
	},
}

type cacheEntryHash = [hashLength * 2]byte // two md5 hashes

// FIXME cache size must be limited, otherwise there will be memory leaks (e.g. with unique RequestIDs in props)
type TemplateRenderer struct {
	cacheMu sync.Mutex
	cache   map[cacheEntryHash][]byte // map key is a unique hash
}

func NewTemplateRenderer() TemplateRenderer {
	return TemplateRenderer{cache: make(map[cacheEntryHash][]byte)}
}

func (tr *TemplateRenderer) Render(content []byte, props Properties) ([]byte, error) {
	if len(content) == 0 {
		return content, nil
	}

	var (
		cacheKey     cacheEntryHash
		cacheKeyInit bool
	)

	if propsHash, err := props.Hash(); err == nil {
		cacheKeyInit, cacheKey = true, tr.mixHashes(propsHash, HashBytes(content))

		tr.cacheMu.Lock()
		item, hit := tr.cache[cacheKey]
		tr.cacheMu.Unlock()

		if hit {
			return item, nil
		}
	}

	var funcMap = template.FuncMap{
		"show_details": func() bool { return props.ShowRequestDetails },
		"hide_details": func() bool { return !props.ShowRequestDetails },
	}

	// make a copy of template functions map
	for s, i := range tplFnMap {
		funcMap[s] = i
	}

	// and allow the direct calling of Properties tokens, e.g. `{{ code | json }}`
	for what, with := range props.Replaces() {
		var n, s = what, with

		funcMap[n] = func() string { return s }
	}

	t, err := template.New("").Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err = t.Execute(&buf, props); err != nil {
		return nil, err
	}

	b := buf.Bytes()

	if cacheKeyInit {
		tr.cacheMu.Lock()
		tr.cache[cacheKey] = b
		tr.cacheMu.Unlock()
	}

	return b, nil
}

func (tr *TemplateRenderer) mixHashes(a, b Hash) (result cacheEntryHash) {
	for i := 0; i < len(a); i++ {
		result[i] = a[i]
	}

	for i := 0; i < len(b); i++ {
		result[i+len(a)] = b[i]
	}

	return
}
