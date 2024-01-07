// A simple custom logger, which checks if the config
// Debug variable is set to true or false, and depending
// on it, logger will log or not log.
package customlogger

import (
	"log"
	"os"

	"github.com/ritchielrez/battery-daemon/config"
)

type CustomLogger struct {
	Info    *log.Logger
	Debug   *log.Logger
	Error   *log.Logger
	Logfile *LogFile
}

type LogFile struct {
	File    *os.File
	Created bool
}

func CustomLoggerInit() *CustomLogger {
	return &CustomLogger{
		Info:  nil,
		Debug: nil,
		Error: nil,
		Logfile: &LogFile{
			File:    nil,
			Created: false,
		},
	}
}

func (cl *CustomLogger) Debugf(format string, args ...interface{}) {
	if config.Debug && cl.Debug == nil {
		cl.Debug = log.New(os.Stdout, "DEBUG: ", log.Ltime)
	} else if !config.Debug {
		return
	}

	cl.Debug.Printf(format, args...)
}

func (cl *CustomLogger) Infof(format string, args ...interface{}) {
	if config.Debug && cl.Info == nil {
		cl.Info = log.New(os.Stdout, "INFO: ", log.Ltime)
	} else if config.LogToFile && cl.Info == nil && cl.Logfile.File == nil {
		var err error
		cl.Logfile.File, err = os.OpenFile(config.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		cl.Info = log.New(cl.Logfile.File, "INFO: ", log.Ltime)
	} else if config.LogToFile && cl.Info == nil && cl.Logfile.File != nil {
	} else {
		return
	}

	cl.Info.Printf(format, args...)
}

func (cl *CustomLogger) Errorf(format string, args ...interface{}) {
	if config.Debug && cl.Info == nil {
		cl.Error = log.New(os.Stdout, "INFO: ", log.Ltime)
	} else if config.LogToFile && cl.Info == nil && cl.Logfile.File == nil {
		var err error
		cl.Logfile.File, err = os.OpenFile(config.LogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error creating file %s, error: %v\n", config.LogFileName, err)
		}
		cl.Info = log.New(cl.Logfile.File, "INFO: ", log.Ltime)
	} else if config.LogToFile && cl.Info == nil && cl.Logfile.File != nil {
		log.Fatalf("Error, expected initial value of Logfile: nil\n")
	} else {
		return
	}

	cl.Info.Printf(format, args...)
}
