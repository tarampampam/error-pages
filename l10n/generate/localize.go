package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"gh.tarampamp.am/error-pages/v4/l10n/generate/jsmin"
)

//go:embed templates/localize.tpl.js
var localizeTpl string

//go:embed templates/playground.tpl.html
var playgroundTpl string

type localesData map[string]map[string]string

func main() {
	var localesFile, jsOutFile, jsMinOutFile, playgroundFile string

	flag.StringVar(&localesFile, "locales", "./locales.json", "input locales file path")
	flag.StringVar(&jsOutFile, "out", "./localize.js", "output JS file path")
	flag.StringVar(&jsMinOutFile, "out-min", "./localize.min.js", "output minified JS file path")
	flag.StringVar(&playgroundFile, "playground", "./playground.html", "output playground file path")
	flag.Parse()

	locales, kErr := readLocalesFile(localesFile)
	exitIfErr(kErr, "parsing locales file")

	minJs, jsErr := writeJSFiles(locales, jsOutFile, jsMinOutFile)
	exitIfErr(jsErr, "writing JS files")

	exitIfErr(writePlaygroundFile(locales, minJs, playgroundFile), "writing playground file")
}

// readLocalesFile reads the specified JSON file and parses it into a map of language codes to translations.
func readLocalesFile(filePath string) (localesData, error) {
	data, rErr := os.ReadFile(filePath)
	if rErr != nil {
		return nil, rErr
	}

	var localesRaw map[string]json.RawMessage
	if err := json.Unmarshal(data, &localesRaw); err != nil {
		return nil, err
	}

	// convert the raw JSON messages into a map of language codes to translations
	locales := make(localesData, len(localesRaw))
	for key, val := range localesRaw {
		var translations map[string]string
		if err := json.Unmarshal(val, &translations); err == nil { // skip non-object values
			locales[key] = translations
		}
	}

	return locales, nil
}

// writeJSFiles generates the JavaScript files for localization based on the provided locales data and writes them
// to the specified paths. It returns the minified JavaScript content as a byte slice or an error if any operation
// fails.
func writeJSFiles(locales localesData, jsPath, jsMinPath string) ([]byte, error) {
	type (
		Translation struct{ LangCode, Value string }
		Token       struct {
			Key          string
			Translations []Translation
		}
	)

	// prepare tokens for the template by converting the locales map into a slice of Token structs
	tokens := make([]Token, 0, len(locales))
	seen := make(map[string]struct{})

	for key, keyTokens := range locales {
		translations := make([]Translation, 0, len(keyTokens))
		for langCode, value := range keyTokens {
			langCode = strings.TrimSpace(strings.ToLower(langCode)) // always trim and lowercase the language code
			translations = append(translations, Translation{LangCode: langCode, Value: value})
			seen[langCode] = struct{}{}
		}

		sort.Slice(translations, func(i, j int) bool { return translations[i].LangCode < translations[j].LangCode })

		tokens = append(tokens, Token{Key: key, Translations: translations})
	}

	sort.Slice(tokens, func(i, j int) bool { return tokens[i].Key < tokens[j].Key })

	supportedLangCodes := slices.Collect(maps.Keys(seen))
	sort.Strings(supportedLangCodes)

	tmpl, tErr := template.New("l10n").Funcs(template.FuncMap{
		"quote": strconv.Quote,
	}).Parse(localizeTpl)
	if tErr != nil {
		return nil, tErr
	}

	var jsBuf bytes.Buffer
	if err := tmpl.Execute(&jsBuf, struct {
		Tokens             []Token
		SupportedLangCodes []string
	}{
		Tokens:             tokens,
		SupportedLangCodes: supportedLangCodes,
	}); err != nil {
		return nil, err
	}

	var jsMinBuf bytes.Buffer
	if err := jsmin.Minify(&jsMinBuf, bytes.NewReader(jsBuf.Bytes())); err != nil {
		return nil, err
	}

	const comment = "// Code generated automatically. DO NOT EDIT. This is a script for the localization of error pages.\n"

	for filePath, buf := range map[string]*bytes.Buffer{
		jsPath:    &jsBuf,
		jsMinPath: &jsMinBuf,
	} {
		f, fErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644) //nolint:mnd
		if fErr != nil {
			return nil, fErr
		}

		if _, err := f.WriteString(comment); err != nil {
			_ = f.Close()

			return nil, err
		}

		if _, err := f.Write(buf.Bytes()); err != nil {
			_ = f.Close()

			return nil, err
		}

		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	return append([]byte(comment), jsMinBuf.Bytes()...), nil
}

// writePlaygroundFile generates and writes the HTML playground file for testing localization based on the provided
// locales data and minified JavaScript content.
func writePlaygroundFile(locales localesData, minJs []byte, filePath string) error {
	tokensList := make([]string, 0, len(locales))
	for token := range locales {
		tokensList = append(tokensList, token)
	}

	sort.Strings(tokensList)

	var (
		langCodes []string
		seen      = make(map[string]struct{})
	)

	for _, translations := range locales {
		for langCode := range translations {
			if _, exists := seen[langCode]; !exists {
				seen[langCode] = struct{}{}
				langCodes = append(langCodes, langCode)
			}
		}
	}

	sort.Strings(langCodes)

	tmpl, tErr := template.New("playground").Funcs(template.FuncMap{
		"quote":      strconv.Quote,
		"l10nScript": func() string { return string(minJs) },
	}).Parse(playgroundTpl)
	if tErr != nil {
		return tErr
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Tokens    []string
		LangCodes []string
	}{
		Tokens:    tokensList,
		LangCodes: langCodes,
	}); err != nil {
		return err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644) //nolint:mnd
	if err != nil {
		return err
	}

	const comment = "<!-- This is a playground for testing the localization of error pages. " +
		"The code is generated. DO NOT EDIT. -->\n"

	_, wErr := f.WriteString(comment)
	if wErr != nil {
		_ = f.Close()

		return wErr
	}

	_, wErr = f.Write(buf.Bytes())
	if wErr != nil {
		_ = f.Close()

		return wErr
	}

	return f.Close()
}

// exitIfErr prints the error message to stderr and exits with status code 1 if err is not nil.
func exitIfErr(err error, act string) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s: %v\n", act, err)

		os.Exit(1)
	}
}
