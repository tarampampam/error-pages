package tpl

import (
	"bytes"
	"sync"

	"github.com/pkg/errors"
)

type (
	// ErrorPages is a error page templates generator.
	ErrorPages struct {
		mu        sync.RWMutex
		templates map[string][]byte            // map[template_name]raw_content
		pages     map[string]*pageProperties   // map[page_code]props
		state     map[string]map[string][]byte // map[template_name]map[page_code]content
	}

	pageProperties struct {
		message, description string
	}
)

var (
	ErrUnknownTemplate = errors.New("unknown template")  // unknown template
	ErrUnknownPageCode = errors.New("unknown page code") // unknown page code
)

// NewErrorPages creates ErrorPages templates generator.
func NewErrorPages() ErrorPages {
	return ErrorPages{
		templates: make(map[string][]byte),
		pages:     make(map[string]*pageProperties),
		state:     make(map[string]map[string][]byte),
	}
}

// AddTemplate to the generator. Template can contain the special placeholders for the error code, message and
// description:
//	{{ code }} - for the code
//	{{ message }} - for the message
//	{{ description }} - for the description
func (e *ErrorPages) AddTemplate(templateName string, content []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.templates[templateName] = content
	e.state[templateName] = make(map[string][]byte)

	for code, props := range e.pages { // update the state
		e.state[templateName][code] = e.makeReplaces(content, code, props.message, props.description)
	}
}

// AddPage with the passed code, message and description. This page will ba available for the all templates.
func (e *ErrorPages) AddPage(code, message, description string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.pages[code] = &pageProperties{message, description}

	for templateName, content := range e.templates { // update the state
		e.state[templateName][code] = e.makeReplaces(content, code, message, description)
	}
}

// GetPage with passed template name and error code.
func (e *ErrorPages) GetPage(templateName, code string) (content []byte, err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if pages, templateExists := e.state[templateName]; templateExists {
		if c, pageExists := pages[code]; pageExists {
			content = c
		} else {
			err = ErrUnknownPageCode
		}
	} else {
		err = ErrUnknownTemplate
	}

	return
}

// IteratePages will call the passed function for each page and template.
func (e *ErrorPages) IteratePages(fn func(template, code string, content []byte) error) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for tplName, codes := range e.state {
		for code, content := range codes {
			if err := fn(tplName, code, content); err != nil {
				return err
			}
		}
	}

	return nil
}

const (
	tknCode byte = iota + 1
	tknMessage
	tknDescription
)

var tknSets = map[byte][][]byte{ //nolint:gochecknoglobals
	tknCode:        {[]byte("{{code}}"), []byte("{{ code }}")},
	tknMessage:     {[]byte("{{message}}"), []byte("{{ message }}")},
	tknDescription: {[]byte("{{description}}"), []byte("{{ description }}")},
}

func (e *ErrorPages) makeReplaces(where []byte, code, message, description string) []byte {
	for tkn, set := range tknSets {
		var replaceWith []byte

		switch tkn {
		case tknCode:
			replaceWith = []byte(code)
		case tknMessage:
			replaceWith = []byte(message)
		case tknDescription:
			replaceWith = []byte(description)
		default:
			panic("tpl: unsupported token") // this is like a fuse, will never occur during normal usage
		}

		if len(replaceWith) > 0 {
			for i := 0; i < len(set); i++ {
				where = bytes.ReplaceAll(where, set[i], replaceWith)
			}
		}
	}

	return where
}
