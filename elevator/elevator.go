package main

import (
	l4g "github.com/alecthomas/log4go"
	elevator "github.com/oleiade/Elevator"
	"log"
)

func main() {
	// Parse command line arguments
	cmdline := &elevator.Cmdline{}
	cmdline.ParseArgs()

    config := elevator.LoadConfig(cmdline)
	// Set up loggers
	l4g.AddFilter("stdout", l4g.INFO, l4g.NewConsoleLogWriter())
    err := elevator.SetupFileLogger("file", config.LogLevel, config.LogFile)
	if err != nil {
		log.Fatal(err)
	}

    exitChannel := elevator.SetupExitChannel()

	if config.Daemon {
		elevator.Daemon(config, exitChannel)
	} else {
		elevator.ListenAndServe(config, exitChannel)
	}
}
