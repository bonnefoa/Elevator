package main

import (
	l4g "github.com/alecthomas/log4go"
	elevator "github.com/oleiade/Elevator"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	exitChannel := make(chan bool)
	go func() {
		sig := <-c
		log.Printf("Received signal '%v', exiting\n", sig)
		exitChannel <- true
	}()

	if config.Daemon {
		elevator.Daemon(config, exitChannel)
	} else {
		elevator.ListenAndServe(config, exitChannel)
	}
}
