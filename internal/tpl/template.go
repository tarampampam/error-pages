package tpl

import (
	"bytes"
	"reflect"
)

type (
	Properties struct {
		Code        string `keyword:"code"`
		Message     string `keyword:"message"`
		Description string `keyword:"description"`
	}
)

func Render(tpl []byte, props Properties) []byte {
	if len(tpl) == 0 {
		return tpl
	}

	var replaces = make(map[string][]byte, reflect.ValueOf(props).NumField())

	for i, v := 0, reflect.ValueOf(props); i < v.NumField(); i++ {
		if keyword, tagExists := v.Type().Field(i).Tag.Lookup("keyword"); tagExists {
			if sv, isString := v.Field(i).Interface().(string); isString && len(sv) > 0 {
				replaces[keyword] = []byte(sv)
			} else {
				replaces[keyword] = []byte{}
			}
		}
	}

	for what, with := range replaces {
		tpl = bytes.ReplaceAll(tpl, []byte("{{"+what+"}}"), with)
		tpl = bytes.ReplaceAll(tpl, []byte("{{ "+what+" }}"), with)
	}

	return tpl
}
