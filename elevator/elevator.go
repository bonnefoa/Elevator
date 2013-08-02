package main

import (
	elevator "github.com/oleiade/Elevator"
)

func main() {
	// Parse command line arguments
	cmdline := &elevator.Cmdline{}
	cmdline.ParseArgs()

    config := elevator.LoadConfig(cmdline)
    elevator.ConfigureLogger(config.LogConfiguration)

    exitChannel := elevator.SetupExitChannel()

	if config.Daemon {
		elevator.Daemon(config, exitChannel)
	} else {
		elevator.ListenAndServe(config, exitChannel)
	}
}
