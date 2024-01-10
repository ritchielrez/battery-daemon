// Package config contains variables that can be changed
// to configure the application's behaviour.
//
// This package is intended to by configured by the user.
package util

import (
	"log"
	"os"
)

// GetConfigDir returns the current user config directory
// path as string.
// Prints an error if there is one and causes
// the program to exit out early.
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
