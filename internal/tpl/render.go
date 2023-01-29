package tpl

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"

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
	"env": os.Getenv,
}

var ErrClosed = errors.New("closed")

type TemplateRenderer struct {
	cacheMu sync.RWMutex
	cache   map[cacheEntryHash]cacheItem // map key is a unique hash

	cacheCleanupInterval time.Duration
	cacheItemLifetime    time.Duration

	close    chan struct{}
	closedMu sync.RWMutex
	closed   bool
}

type (
	cacheEntryHash = [hashLength * 2]byte // two md5 hashes
	cacheItem      struct {
		data          []byte
		expiresAtNano int64
	}
)

const (
	cacheCleanupInterval = time.Second
	cacheItemLifetime    = time.Second * 2
)

// NewTemplateRenderer returns new template renderer. Don't forget to call Close() function!
func NewTemplateRenderer() *TemplateRenderer {
	tr := &TemplateRenderer{
		cache:                make(map[cacheEntryHash]cacheItem),
		cacheCleanupInterval: cacheCleanupInterval,
		cacheItemLifetime:    cacheItemLifetime,
		close:                make(chan struct{}, 1),
	}

	go tr.cleanup()

	return tr
}

func (tr *TemplateRenderer) cleanup() {
	defer close(tr.close)

	timer := time.NewTimer(tr.cacheCleanupInterval)
	defer timer.Stop()

	for {
		select {
		case <-tr.close:
			tr.cacheMu.Lock()
			for hash := range tr.cache {
				delete(tr.cache, hash)
			}
			tr.cacheMu.Unlock()

			return

		case <-timer.C:
			tr.cacheMu.Lock()
			var now = time.Now().UnixNano()

			for hash, item := range tr.cache {
				if now > item.expiresAtNano {
					delete(tr.cache, hash)
				}
			}
			tr.cacheMu.Unlock()

			timer.Reset(tr.cacheCleanupInterval)
		}
	}
}

func (tr *TemplateRenderer) Render(content []byte, props Properties) ([]byte, error) { //nolint:funlen
	if tr.isClosed() {
		return nil, ErrClosed
	}

	if len(content) == 0 {
		return content, nil
	}

	var (
		cacheKey     cacheEntryHash
		cacheKeyInit bool
	)

	if propsHash, err := props.Hash(); err == nil {
		cacheKeyInit, cacheKey = true, tr.mixHashes(propsHash, HashBytes(content))

		tr.cacheMu.RLock()
		item, hit := tr.cache[cacheKey]
		tr.cacheMu.RUnlock()

		if hit {
			// cache item has been expired?
			if time.Now().UnixNano() > item.expiresAtNano {
				tr.cacheMu.Lock()
				delete(tr.cache, cacheKey)
				tr.cacheMu.Unlock()
			} else {
				return item.data, nil
			}
		}
	}

	var funcMap = template.FuncMap{
		"show_details":  func() bool { return props.ShowRequestDetails },
		"hide_details":  func() bool { return !props.ShowRequestDetails },
		"l10n_disabled": func() bool { return props.L10nDisabled },
		"l10n_enabled":  func() bool { return !props.L10nDisabled },
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
		tr.cache[cacheKey] = cacheItem{
			data:          b,
			expiresAtNano: time.Now().UnixNano() + tr.cacheItemLifetime.Nanoseconds(),
		}
		tr.cacheMu.Unlock()
	}

	return b, nil
}

func (tr *TemplateRenderer) isClosed() (closed bool) {
	tr.closedMu.RLock()
	closed = tr.closed
	tr.closedMu.RUnlock()

	return
}

func (tr *TemplateRenderer) Close() error {
	if tr.isClosed() {
		return ErrClosed
	}

	tr.closedMu.Lock()
	tr.closed = true
	tr.closedMu.Unlock()

	tr.close <- struct{}{}

	return nil
}

func (tr *TemplateRenderer) mixHashes(a, b Hash) (result cacheEntryHash) {
	for i := 0; i < len(a); i++ { //nolint:gosimple
		result[i] = a[i]
	}

	for i := 0; i < len(b); i++ {
		result[i+len(a)] = b[i]
	}

	return
}
