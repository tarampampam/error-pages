package tpl

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
	"gh.tarampamp.am/error-pages/v4/l10n"
)

var fns = template.FuncMap{ //nolint:gochecknoglobals
	// returns the current time:
	//	`{{ now.Format "2006-01-02 15:04:05" }}`	// `2024-06-01 12:34:56`
	//	`{{ now.Unix }}`													// `1712128496`
	"now": time.Now,

	// returns the hostname of the machine running the application:
	//	`{{ hostname }}`	// `my-hostname`
	"hostname": hostname,

	// returns the JSON representation of the given value:
	//	`{{ toJson hostname }}`	// `"my-hostname"`
	//	`{{ "test" | toJson }}`	// `"test"`
	//	`{{ 42 | toJson }}`			// `42`
	"toJson": toJSON,
	"toJSON": toJSON, // is an alias for toJson, as some users may prefer this naming convention

	// returns the integer representation of the given value, or 0 if it cannot be converted:
	//	`{{ "42" | int }}`			// `42`
	//	`{{ int 42 }}`					// `42`
	//	`{{ int 3.14 }}`				// `3`
	//	`{{ int "test" }}`			// `0`
	//	`{{ "42test" | int }}`	// `0`
	//	`{{ now.Unix | int }}`	// `1712128496`
	"toInt": toInt,
	"int":   toInt, // is an alias for toInt, as some users may prefer this shorter name

	// returns the current application version:
	//	`{{ version }}`	// `1.0.0`
	"version": appVersion,

	// returns the value of the specified environment variable, or an empty string if it is not set. For security reasons,
	// some environment variables (those containing "PASSWORD", "SECRET", "KEY" and others) will return a string of
	// asterisks (*) instead of the actual value:
	//	`{{ env "SHELL" }}`	// `/bin/bash`
	"env": getEnv,

	// returns the escaped string, safe for HTML contexts. It replaces special characters with their
	// corresponding HTML entities:
	//	`{{ "<test>" | escape }}`	// `&lt;test&gt;`
	"escape": html.EscapeString,

	// returns trimmed string with leading and trailing whitespace removed:
	//	`{{ "  test  " | trim }}`	// `test`
	"trim": strings.TrimSpace,

	// returns the string with the specified prefix removed, if it exists:
	//	`{{ "test" | trimPrefix "te" }}`	// `st`
	"trimPrefix": trimPrefix,

	// returns the string with the specified suffix removed, if it exists:
	//	`{{ "test" | trimSuffix "st" }}`	// `te`
	//	`{{ trimSuffix "st" "test" }}`	// `te`
	//	`{{ "test" | trimPostfix "st" }}`	// `te`
	"trimSuffix":  trimSuffix,
	"trimPostfix": trimSuffix, // is an alias for trimSuffix, as "postfix" is a more intuitive term for some users

	// returns the string with all occurrences of the old substring replaced by the new substring:
	//	`{{ "test" | replace "t" "z" }}`	// `zesz`
	"replace": replace,

	// returns true if the source string contains the specified substring:
	//	`{{ "test" | contains "es" }}`	// `true`
	"contains": contains,

	// returns the number of non-overlapping occurrences of the substring in the source string:
	//	`{{ "test" | count "t" }}`	// `2`
	"count": count,

	// returns the fields (words) in the source string, split by whitespace:
	//	`{{ "foo bar baz" | fields }}`	// `["foo", "bar", "baz"]`
	"fields": strings.Fields,

	// returns the lowercase version of the string:
	//	`{{ "TEST" | lower }}`	// `test`
	"lower": strings.ToLower,

	// returns the uppercase version of the string:
	//	`{{ "test" | upper }}`	// `TEST`
	"upper": strings.ToUpper,

	// returns the given value if it is set, or the default value if it is not set:
	//	`{{ env "NOT_SET" | default "default value" }}`	// `default value`
	//	`{{ .OriginalURI | default "N/A" }}`						// `N/A`
	"default": def,

	// returns true if the source string starts with the specified prefix:
	//	`{{ "test" | hasPrefix "te" }}`	// `true`
	"hasPrefix": hasPrefix,

	// returns true if the source string ends with the specified suffix:
	//	`{{ "test" | hasSuffix "st" }}`	// `true`
	"hasSuffix":  hasSuffix,
	"hasPostfix": hasSuffix, // is an alias for hasSuffix, as "postfix" is a more intuitive term for some users

	// joins the elements of a slice into a single string separated by sep:
	//	`{{ split "," "a,b,c" | join ", " }}`		// `a, b, c`
	//	`{{ fields "foo bar baz" | join "-" }}`	// `foo-bar-baz`
	//	`{{ join ", " (split "," "a,b,c") }}`		// `a, b, c`
	"join": join,

	// splits the source string into a slice of substrings by the given separator:
	//	`{{ "a,b,c" | split "," }}`                      // `[a b c]`
	//	`{{ range "a,b,c" | split "," }}<li>{{.}}</li>{{ end }}`
	"split": split,

	// wraps the string in double quotes, escaping special characters using Go string literal rules:
	//	`{{ "test" | quote }}`					// `"test"`
	//	`{{ "it's \"fine\"" | quote }}`	// `"it's \"fine\""`
	"quote": strconv.Quote,

	// wraps the string in single quotes:
	//	`{{ "test" | squote }}`	// `'test'`
	"squote": squote,

	// returns a string consisting of count copies of the input string:
	//	`{{ "Ha" | repeat 3 }}`	// `HaHaHa`
	"repeat": repeat,

	// returns the substring of s starting at index start with the given length. If start is negative, it defaults to 0.
	// If length is negative, the substring extends to the end of s. If start+length exceeds len(s), it is clamped:
	//	`{{ "test" | substr 1 4 }}`						// `est`
	//	`{{ "test" | substr -1 2 }}`					// `te`
	//	`{{ "test" | substr 2 -1 }}`					// `st`
	//	`{{ "Hello, World!" | substr 7 5 }}`	// `World`
	"substr": substr,

	// converts any value to its string representation:
	//	`{{ .StatusCode | toString }}`	// `404`
	//	`{{ 42 | toString }}`						// `42`
	"toString": toString,
	"str":      toString, // is an alias for toString, as some users may prefer this shorter name

	// returns trueVal if condition is true, falseVal otherwise:
	//	`{{ ternary "shown" "hidden" .Config.ShowRequestDetails }}`		// `shown`
	//	`{{ .Config.ShowRequestDetails | ternary "shown" "hidden" }}`	// `shown`
	"ternary": ternary,

	// returns the first non-empty value from the given list, or an empty string if all are empty:
	//	`{{ coalesce .Message .Description "Unknown error" }}`	// first non-empty
	"coalesce": coalesce,

	// returns the URL query-escaped form of the string:
	//	`{{ .OriginalURI | urlEncode }}`	// `/api/users` → `%2Fapi%2Fusers`
	"urlEncode": url.QueryEscape,

	// returns true if the given value is empty (zero value for its type):
	//	`{{ if isEmpty .Message }}No message{{ end }}`
	"isEmpty": empty,

	// returns true if the given value is not empty:
	//	`{{ if isNotEmpty .Message }}{{ .Message }}{{ end }}`
	"isNotEmpty": isNotEmpty,

	// returns the string truncated to at most n runes; appends "..." if the string was truncated:
	//	`{{ .Description | truncate 120 }}`	// `Some long description...`
	"truncate": truncate,

	// returns the string with all leading and trailing occurrences of any character in cutset removed:
	//	`{{ ".....test....." | trimAll "." }}`	// `test`
	//	`{{ "!?test?!" | trimAll "!?" }}`				// `test`
	"trimAll": trimAll,

	// returns the content of the JS file with a script for automatic error page localization:
	//	`{{ l10nScript }}`	// `Object.defineProperty(window, ...`
	"l10nScript": l10n.L10n,

	// Deprecated: use `{{ now.Unix }}` instead.
	"nowUnix": nowUnix,
	// Deprecated: use `{{ "test" | count "t" }}` instead.
	"strCount": strings.Count,
	// Deprecated: use `{{ "test" | contains "es" }}` instead.
	"strContains": strings.Contains,
	// Deprecated: use `{{ "  test  " | trim }}` instead.
	"strTrimSpace": strings.TrimSpace,
	// Deprecated: use `{{ "test" | trimPrefix "te" }}` instead.
	"strTrimPrefix": strings.TrimPrefix,
	// Deprecated: use `{{ "test" | trimSuffix "st" }}` instead.
	"strTrimSuffix": strings.TrimSuffix,
	// Deprecated: use `{{ "test" | replace "t" "z" }}` instead.
	"strReplace": strings.ReplaceAll,
	// Deprecated: kept for backward compatibility.
	"strIndex": strings.Index,
	// Deprecated: use `{{ "foo bar baz" | fields }}` instead.
	"strFields": strings.Fields,
	// Deprecated: use `{{ "test" | toJson }}` instead.
	"json": toJSON,
}

// nowUnix returns the current time in Unix format (seconds since 1970 UTC).
//
// Deprecated: use `{{ now.Unix }}` instead.
func nowUnix() int { return int(time.Now().Unix()) }

// hostname returns the hostname of the machine running the application, or an empty string if it cannot be determined.
func hostname() (hostname string) {
	hostname, _ = os.Hostname() //nolint:errcheck

	return
}

// toJSON is a helper function that converts any value to its JSON string representation. It ignores any errors during
// marshaling, returning an empty string if the conversion fails.
func toJSON(v any) string {
	b, _ := json.Marshal(v) //nolint:errcheck,errchkjson

	return string(b)
}

// toInt attempts to convert any value to an int, returning 0 if the conversion is not possible.
//
// Hot path covers all primitive types (int*, uint*, float*, complex*, bool, string) without reflection. For complex
// types only the real part is used. For type aliases and pointer types reflection is used as a fallback.
//
// Note: if a type implements [fmt.Stringer], the string representation takes priority over the underlying numeric
// kind. If String() returns a non-numeric value, the result is 0.
func toInt(value any) int { //nolint:funlen
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v) //nolint:gosec
	case float32:
		return int(v)
	case float64:
		return int(v)
	case complex64:
		return int(real(v))
	case complex128:
		return int(real(v))
	case bool:
		if v {
			return 1
		}

		return 0
	case nil:
		return 0
	case string:
		return stringToInt(v)
	case fmt.Stringer:
		return stringToInt(v.String())
	default:
		rv := reflect.ValueOf(v)

		for rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return 0
			}

			rv = rv.Elem()
		}

		switch rv.Kind() { //nolint:exhaustive
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(rv.Uint()) //nolint:gosec
		case reflect.Float32, reflect.Float64:
			return int(rv.Float())
		case reflect.Complex64, reflect.Complex128:
			return int(real(rv.Complex()))
		case reflect.Bool:
			if rv.Bool() {
				return 1
			}

			return 0
		case reflect.String:
			return stringToInt(rv.String())
		default:
			return 0
		}
	}
}

// stringToInt tries Atoi first (fast path), then ParseFloat for values like "3.14". Input is trimmed before parsing.
func stringToInt(s string) int {
	s = strings.TrimSpace(s)

	if n, err := strconv.Atoi(s); err == nil {
		return n
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}

	return 0
}

// appVersion returns the current application version as a string.
func appVersion() string { return appmeta.Version() }

// getEnv retrieves the value of the environment variable named by the key. If the variable is not present, it
// returns an empty string.
//
// For security reasons, if the key contains any of the following substrings (case-insensitive): "PASSWORD", "SECRET",
// "KEY", "TOKEN", "PASS", "PWD", "CRED", the function returns a string of asterisks (*) with the same length as the
// actual value, instead of the value itself.
func getEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return ""
	}

	for segment := range strings.SplitSeq(strings.ToUpper(key), "_") {
		switch segment {
		case "PASSWORD", "SECRET", "KEY", "TOKEN", "PASS", "PWD", "CRED":
			return strings.Repeat("*", utf8.RuneCountInString(v))
		}
	}

	return v
}

// trimPrefix is a helper function that removes the specified prefix from the given string.
func trimPrefix(s, src string) string { return strings.TrimPrefix(src, s) }

// trimSuffix is a helper function that removes the specified suffix from the given string.
func trimSuffix(s, src string) string { return strings.TrimSuffix(src, s) }

// replace is a helper function that replaces all occurrences of the old substring with the new substring in the given
// string.
func replace(old, replacement, src string) string { return strings.ReplaceAll(src, old, replacement) }

// contains is a helper function that checks if the given substring is present in the source string.
func contains(substr, src string) bool { return strings.Contains(src, substr) }

// count is a helper function that counts the number of non-overlapping occurrences of the substring in the
// source string.
func count(substr, src string) int { return strings.Count(src, substr) }

// def checks whether `given` is set, and returns default if not.
//
// This returns `d` if `given` appears not to be set, and `given` otherwise.
//
// For numeric types 0 is unset.
// For strings, maps, arrays, and slices, len() = 0 is considered unset.
// For bool, false is unset.
// Structs are never considered unset.
//
// For everything else, including pointers, a nil value is unset.
func def(d any, given ...any) any {
	if empty(given) || empty(given[0]) {
		return d
	}

	return given[0]
}

// empty returns true if the given value has the zero value for its type.
func empty(given any) bool { //nolint:funlen
	switch v := given.(type) {
	case nil:
		return true
	case bool:
		return !v
	case string:
		return v == ""
	case int:
		return v == 0
	case int8:
		return v == 0
	case int16:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case uint:
		return v == 0
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0
	case uintptr:
		return v == 0
	case float32:
		return v == 0
	case float64:
		return v == 0
	case complex64:
		return v == 0
	case complex128:
		return v == 0
	default:
		g := reflect.ValueOf(given)
		if !g.IsValid() {
			return true
		}

		for g.Kind() == reflect.Pointer {
			if g.IsNil() {
				return true
			}

			g = g.Elem()
		}

		switch g.Kind() { //nolint:exhaustive
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return g.Len() == 0
		case reflect.Bool:
			return !g.Bool()
		case reflect.Complex64, reflect.Complex128:
			return g.Complex() == 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return g.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return g.Uint() == 0
		case reflect.Float32, reflect.Float64:
			return g.Float() == 0
		case reflect.Struct:
			return false
		default:
			return g.IsNil()
		}
	}
}

// hasPrefix is a helper function that checks if the source string starts with the specified prefix.
func hasPrefix(s, src string) bool { return strings.HasPrefix(src, s) }

// hasSuffix is a helper function that checks if the source string ends with the specified suffix.
func hasSuffix(s, src string) bool { return strings.HasSuffix(src, s) }

// join concatenates the elements of a slice into a single string separated by sep. Non-string elements are converted
// using [fmt.Sprint]. Handles []string and []any directly; falls back to reflection for other slice/array types.
func join(sep string, v any) string {
	switch s := v.(type) {
	case []string:
		return strings.Join(s, sep)
	case []any:
		strs := make([]string, len(s))
		for i, item := range s {
			strs[i] = fmt.Sprint(item)
		}

		return strings.Join(strs, sep)
	default:
		rv := reflect.ValueOf(v)

		k := rv.Kind()
		if k != reflect.Slice && k != reflect.Array {
			return fmt.Sprint(v)
		}

		strs := make([]string, rv.Len())
		for i := range rv.Len() {
			strs[i] = fmt.Sprint(rv.Index(i).Interface())
		}

		return strings.Join(strs, sep)
	}
}

// split splits src into a slice of substrings separated by sep.
func split(sep, src string) []string { return strings.Split(src, sep) }

// squote wraps a string in single quotes.
func squote(s string) string { return "'" + s + "'" }

// repeat returns a string consisting of count copies of the input string.
func repeat(count int, s string) string { return strings.Repeat(s, count) }

// substr returns the substring of s starting at rune index start with the given rune length. If start is negative,
// it defaults to 0. If length is negative, the substring extends to the end of s. If start+length exceeds the rune
// count of s, it is clamped.
func substr(start, length int, s string) string {
	runes := []rune(s)
	n := len(runes)

	if start < 0 {
		start = 0
	}

	if start >= n {
		return ""
	}

	if length < 0 {
		return string(runes[start:])
	}

	end := min(start+length, n)

	return string(runes[start:end])
}

// toString converts any value to its string representation.
func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	case int:
		return strconv.Itoa(s)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int32:
		return strconv.FormatInt(int64(s), 10)
	case int64:
		return strconv.FormatInt(s, 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(s, 10)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(s)
	case nil:
		return ""
	case fmt.Stringer:
		return s.String()
	default:
		rv := reflect.ValueOf(v)

		for rv.Kind() == reflect.Pointer {
			if rv.IsNil() {
				return ""
			}

			rv = rv.Elem()
		}

		switch rv.Kind() { //nolint:exhaustive
		case reflect.String:
			return rv.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(rv.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(rv.Uint(), 10)
		case reflect.Float32:
			return strconv.FormatFloat(rv.Float(), 'f', -1, 32)
		case reflect.Float64:
			return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
		case reflect.Bool:
			return strconv.FormatBool(rv.Bool())
		case reflect.Slice:
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				return string(rv.Bytes())
			}

			return fmt.Sprint(rv.Interface())
		default:
			return fmt.Sprint(rv.Interface())
		}
	}
}

// ternary returns trueVal if condition is true, falseVal otherwise.
func ternary(trueVal, falseVal any, condition bool) any {
	if condition {
		return trueVal
	}

	return falseVal
}

// coalesce returns the first non-empty value from the given list, or "" (empty string) if all values are empty.
func coalesce(vals ...any) any {
	for _, v := range vals {
		if !empty(v) {
			return v
		}
	}

	return ""
}

// isNotEmpty returns true if the given value is not empty.
func isNotEmpty(v any) bool { return !empty(v) }

// truncate returns the string truncated to at most n runes. If truncated, "..." is appended so the total length
// is at most n runes.
func truncate(n int, s string) string {
	const ellipsis = "..."

	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	if n < len(ellipsis) {
		return string(runes[:n])
	}

	return string(runes[:n-len(ellipsis)]) + ellipsis
}

// trimAll returns the string with all leading and trailing occurrences of any character in cutset removed.
func trimAll(cutset, s string) string { return strings.Trim(s, cutset) }
