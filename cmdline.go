package elevator

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"log"
)

type Cmdline struct {
	ConfigFile *string
	DaemonMode *bool
	Endpoint   *string
	LogLevel   *string
}

func (c *Cmdline) ParseArgs() {
	c.ConfigFile = flag.String("c",
		DEFAULT_CONFIG_FILE,
		"Specifies config file path")
	c.DaemonMode = flag.Bool("d",
		DEFAULT_DAEMON_MODE,
		"Launches elevator as a daemon")
	c.Endpoint = flag.String("e",
		DEFAULT_ENDPOINT,
		"Endpoint to bind elevator to")
	c.LogLevel = flag.String("l",
		DEFAULT_LOG_LEVEL,
		"Sets elevator verbosity")
	flag.Parse()
}

func SetupExitChannel() chan bool {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	exitChannel := make(chan bool)
	go func() {
		sig := <-c
		log.Printf("Received signal '%v', exiting\n", sig)
		exitChannel <- true
	}()
    return exitChannel
}
