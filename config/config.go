// This package contains variables that can be changed
// to configure the application's behaviour.
package config

import "os"

var USER_HOME = os.Getenv("HOME")

const (
	Debug     = false
	LogToFile = true // Do not enable debug mode to make this work
)

var (
	LogFileName = USER_HOME + "/.config/battery-daemon/battery-daemon.log" // Full path is needed
)
