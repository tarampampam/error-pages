package tpl

import "reflect"

type Properties struct {
	Code        string `token:"code"`
	Message     string `token:"message"`
	Description string `token:"description"`
}

// Replaces return a map with strings for the replacing, where the map key is a token.
func (p *Properties) Replaces() map[string]string {
	var replaces = make(map[string]string, reflect.ValueOf(*p).NumField())

	for i, v := 0, reflect.ValueOf(*p); i < v.NumField(); i++ {
		if keyword, tagExists := v.Type().Field(i).Tag.Lookup("token"); tagExists {
			if sv, isString := v.Field(i).Interface().(string); isString && len(sv) > 0 {
				replaces[keyword] = sv
			} else {
				replaces[keyword] = ""
			}
		}
	}

	return replaces
}
