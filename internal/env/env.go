// Package env contains all about environment variables, that can be used by current application.
package env

import "os"

type envVariable string

const (
	ListenAddr     envVariable = "LISTEN_ADDR"   // IP address for listening
	ListenPort     envVariable = "LISTEN_PORT"   // port number for listening
	TemplateName   envVariable = "TEMPLATE_NAME" // template name
	ConfigFilePath envVariable = "CONFIG_FILE"   // path to the config file
)

// String returns environment variable name in the string representation.
func (e envVariable) String() string { return string(e) }

// Lookup retrieves the value of the environment variable. If the variable is present in the environment the value
// (which may be empty) is returned and the boolean is true. Otherwise the returned value will be empty and the
// boolean will be false.
func (e envVariable) Lookup() (string, bool) { return os.LookupEnv(string(e)) }
