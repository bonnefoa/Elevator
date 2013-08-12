package server

import (
	l4g "github.com/alecthomas/log4go"
	"os"
)

func (c *LogConfiguration) L4gLevel() l4g.Level {
	return logLevels[c.LogLevel]
}

// Log levels binding
var logLevels = map[string]l4g.Level{
	l4g.DEBUG.String():    l4g.DEBUG,
	l4g.FINEST.String():   l4g.FINEST,
	l4g.FINE.String():     l4g.FINE,
	l4g.DEBUG.String():    l4g.DEBUG,
	l4g.TRACE.String():    l4g.TRACE,
	l4g.INFO.String():     l4g.INFO,
	l4g.WARNING.String():  l4g.WARNING,
	l4g.ERROR.String():    l4g.ERROR,
	l4g.CRITICAL.String(): l4g.CRITICAL,
}

// SetupLogger function ensures logging file exists, and
// is writable, and sets up a log4go filter accordingly
func addFilterFile(logger_name string, c *LogConfiguration) error {
	// Check file exists or return the error
	if _, err := os.Stat(c.LogFile); err != nil {
		return err
	}

	// check file permissions are correct
	_, err := os.OpenFile(c.LogFile, os.O_WRONLY, 0400)
	if err != nil {
		return err
	}

	l4g.AddFilter(logger_name, c.L4gLevel(),
		l4g.NewFileLogWriter(c.LogFile, false))

	return nil
}

func ConfigureLogger(c *LogConfiguration) error {
	// Set up loggers
	l4g.AddFilter("stdout", c.L4gLevel(), l4g.NewConsoleLogWriter())
	err := addFilterFile("file", c)
	return err
}
