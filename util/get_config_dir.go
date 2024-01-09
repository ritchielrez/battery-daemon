// This package contains simple functions,  which may be
// needed by other packages.

// To configure this small application, there os a `config.go`
// file, where the user needs to specify the config directory.
// In linux, configs are usually stored in ~/.config/ directory.
// But it may be different in certain systems. Thus providing a
// simple function to find the user specific config directory.
package util

import (
	"log"
	"os"
)

func GetConfigDir() string {
	var val string
	var ok bool
	val, ok = os.LookupEnv("XDG_CONFIG_HOME")

	if !ok {
		val, ok = os.LookupEnv("HOME")
		if !ok {
			log.Fatalf("Error: could not fine $XDG_CONFIG_HOME or $HOME\n")
		}

		return val + "/.config/"
	}

	return val + "/"
}
