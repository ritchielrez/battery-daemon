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

func (logFile *LogFile) LogFileCreate() {
	var err error

	logFile.Created = true

	logFile.File, err = os.OpenFile(
		config.LogFileName,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
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
		cl.Debug = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Ldate)
	} else if !config.Debug {
		return
	}

	cl.Debug.Printf(format, args...)
}

func (cl *CustomLogger) Infof(format string, args ...interface{}) {
	if config.Debug && cl.Info == nil {
		cl.Info = log.New(os.Stdout, "INFO: ", log.Ltime|log.Ldate)
	} else if config.LogToFile && cl.Info == nil && !cl.Logfile.Created && cl.Logfile.File == nil {
		cl.Logfile.LogFileCreate()
		cl.Info = log.New(cl.Logfile.File, "INFO: ", log.Ltime|log.Ldate)
	} else if config.LogToFile && cl.Info == nil && !cl.Logfile.Created && cl.Logfile.File != nil {
		log.Fatalf("Error: Customlogger.LogFile.File struct was not created by LogFileCreate()\n")
	} else if config.LogToFile && cl.Info == nil && cl.Logfile.Created && cl.Logfile.File != nil {
		cl.Info = log.New(cl.Logfile.File, "INFO: ", log.Ltime|log.Ldate)
	}

	cl.Info.Printf(format, args...)
}

func (cl *CustomLogger) Errorf(format string, args ...interface{}) {
	if config.Debug && cl.Error == nil {
		cl.Error = log.New(os.Stdout, "ERROR: ", log.Ltime|log.Ldate)
	} else if config.LogToFile && cl.Error == nil && !cl.Logfile.Created && cl.Logfile.File == nil {
		cl.Logfile.LogFileCreate()
		cl.Error = log.New(cl.Logfile.File, "ERROR: ", log.Ltime|log.Ldate)
	} else if config.LogToFile && cl.Error == nil && !cl.Logfile.Created && cl.Logfile.File != nil {
		log.Fatalf("Error: Customlogger.LogFile.File struct was not created by LogFileCreate()\n")
	} else if config.LogToFile && cl.Error == nil && cl.Logfile.Created && cl.Logfile.File != nil {
		cl.Error = log.New(cl.Logfile.File, "ERROR: ", log.Ltime|log.Ldate)
	}

	cl.Error.Fatalf(format, args...)
}
