// Package customlogger implements a simple custom logger,
// for logging to file or stdout depending on config package.
//
// This package is only intended to be used with this project.
package customlogger

import (
	"log"
	"os"

	"github.com/ritchielrez/battery-daemon/config"
)

// A CustomLogger stores dirrent types of loggers in a single struct.
type CustomLogger struct {
	Info    *log.Logger
	Debug   *log.Logger
	Error   *log.Logger
	Logfile *LogFile
}

// A LogFile stores information if the file was created or opened and
// the *os.File itself.
type LogFile struct {
	File    *os.File
	Created bool
}

// LogFileCreate creates or opens a logfile and makes to store that
// information in the Logfile struct.
// Prints an error if there is one and causes
// the program to exit out early.
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

// CustomLoggerInit initializes a CustomLogger with proper values
// that is expected by other functions in the package. Returns
// the initialized CustomLogger as a pointer.
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

// Debugf prints the debug msg in the stdout, if only config.Debug is true.
// Arguments are the same as `fmt.Printf()`.
func (cl *CustomLogger) Debugf(format string, args ...interface{}) {
	if config.Debug && cl.Debug == nil {
		cl.Debug = log.New(os.Stdout, "DEBUG: ", log.Ltime|log.Ldate)
	} else if !config.Debug {
		return
	}

	cl.Debug.Printf(format, args...)
}

// Infof prints to a logfile or stdout, depending on the bool values set in
// the config package.
// Arguments are the same as `fmt.Printf()`.
// Prints an error if there is one and causes
// the program to exit out early.
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

// Infof prints to a logfile or stdout, depending on the bool values set in
// the config package. Also causes the program to exit out with return code 1.
// Arguments are the same as `fmt.Printf()`.
// Prints an error if there is one and causes
// the program to exit out early.
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
