package tpl

import (
	"errors"
	"sync"
)

// Annotator allows to annotate error code.
type Annotator struct {
	Message     string
	Description string
}

// Errors is a "cached storage" for the rendered error pages for the different templates and codes.
type Errors struct {
	templates map[string][]byte
	codes     map[string]Annotator

	cacheMu sync.RWMutex
	cache   map[string]map[string][]byte // map[template]map[code]content
}

// NewErrors creates new Errors.
func NewErrors(templates map[string][]byte, codes map[string]Annotator) *Errors {
	return &Errors{
		templates: templates,
		codes:     codes,
		cache:     make(map[string]map[string][]byte),
	}
}

func (e *Errors) existsInCache(template, code string) ([]byte, bool) {
	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()

	if codes, tplOk := e.cache[template]; tplOk {
		if content, codeOk := codes[code]; codeOk {
			return content, true
		}
	}

	return nil, false
}

func (e *Errors) putInCache(template, code string) error {
	if _, ok := e.templates[template]; !ok {
		return errors.New("template \"" + template + "\" does not exists")
	}

	if _, ok := e.codes[code]; !ok {
		return errors.New("code \"" + code + "\" does not exists")
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	if _, ok := e.cache[template]; !ok {
		e.cache[template] = make(map[string][]byte)
	}

	e.cache[template][code] = Replace(e.templates[template], Replaces{
		Code:        code,
		Message:     e.codes[code].Message,
		Description: e.codes[code].Description,
	})

	return nil
}

// Get the rendered error page content.
func (e *Errors) Get(template, code string) ([]byte, error) {
	if content, ok := e.existsInCache(template, code); ok {
		return content, nil
	}

	if err := e.putInCache(template, code); err != nil {
		return nil, err
	}

	e.cacheMu.RLock()
	defer e.cacheMu.RUnlock()

	return e.cache[template][code], nil
}

// VisitAll allows to iterate all possible error pages and templates.
func (e *Errors) VisitAll(fn func(template, code string, content []byte) error) error {
	for tpl := range e.templates {
		for code := range e.codes {
			content, err := e.Get(tpl, code)
			if err != nil {
				return err // will never happen
			}

			if err = fn(tpl, code, content); err != nil {
				return err
			}
		}
	}

	return nil
}
