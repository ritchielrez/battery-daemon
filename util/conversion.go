// Package config contains variables that can be changed
// to configure the application's behaviour.
//
// This package is intended to by configured by the user.
package util

import (
	"log"
	"strconv"
)

// Converts a string to int. Returns the int.
// Prints an error if there is one and causes
// the program to exit out early.
func StringToInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Cannot convert string %s to integer, error: %v\n", s, err)
	}
	return num
}
