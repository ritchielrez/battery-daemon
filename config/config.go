// Package config contains variables that can be changed
// to configure the application's behaviour.
//
// This package is intended to by configured by the user.
package config

import "github.com/ritchielrez/battery-daemon/internal/util"

const (
	Debug     = false
	LogToFile = true // Do not enable debug mode to make this work
)

var (
	LogFileName = util.GetConfigDir() + "battery-daemon/battery-daemon.log" // Full path is needed
)
