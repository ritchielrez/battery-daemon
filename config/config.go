// This package contains variables that can be changed
// to configure the application's behaviour.
package config

import "github.com/ritchielrez/battery-daemon/util"

const (
	Debug     = false
	LogToFile = true // Do not enable debug mode to make this work
)

var (
	LogFileName = util.GetConfigDir() + "battery-daemon/battery-daemon.log" // Full path is needed
)
