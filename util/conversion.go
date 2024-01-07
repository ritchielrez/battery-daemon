// This package contains simple functions,  which may be
// needed by other packages.
package util

import (
	"log"
	"strconv"
)

func StringToInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
        log.Fatalf("Cannot convert string %s to integer, error: %v\n", s, err)
	}
	return num
}
