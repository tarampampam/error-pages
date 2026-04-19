package appmeta

import "strings"

// version value should be set at build time using -ldflags, for example:
//
//	go build -ldflags "-X <this-package-path>/appmeta.version=${APP_VERSION}" ...
var version = "0.0.0@undefined"

// Version returns version value (without `v` prefix).
func Version() string { return stripVersionPrefix(strings.TrimSpace(version)) }
