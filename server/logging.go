package server

import (
)

// SetupLogger function ensures logging file exists, and
// is writable, and sets up a log4go filter accordingly
func addFilterFile(loggerName string, c *logConfiguration) error {
	return nil
}

func configureLogger(c *logConfiguration) error {
	// Set up loggers
	//l4g.AddFilter("stdout", c.L4gLevel(), l4g.NewConsoleLogWriter())
	//err := addFilterFile("file", c)
	//return err
	return nil
}
