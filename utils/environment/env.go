package environment

import (
	"os"
	"strings"
)

// GetOrElse gets an environment variable defined in envvar or returns a
// default value
func GetOrElse(envvar, value string) string {
	if env := os.Getenv(strings.Trim(envvar, " ")); env != "" {
		return env
	}
	return value
}
