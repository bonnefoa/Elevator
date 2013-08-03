package main

import (
	"flag"
	elevator "github.com/oleiade/Elevator"
	"log"
	"os"
)

func checkErr(config *elevator.Config, fs *flag.FlagSet, err error) {
	if err != nil {
		fs.PrintDefaults()
		if config != nil {
			log.Printf("Current config is %s", config)
		}
		log.Fatal(err)
	}
}

func main() {
	fs := flag.NewFlagSet("Elevator flag set", flag.ContinueOnError)

	conf_file := fs.String("c",
		elevator.DEFAULT_CONFIG_FILE,
		"Specifies config file path")
	err := fs.Parse(os.Args[1:])
	checkErr(nil, fs, err)

	config, err := elevator.ConfFromFile(*conf_file)
	elevator.SetFlag(fs, config.CoreConfig)
	elevator.SetFlag(fs, config.LogConfiguration)
	checkErr(config, fs, err)

	// Reparse to make command line arguments override conf file
	err = fs.Parse(os.Args[:1])
	checkErr(config, fs, err)

	err = elevator.ConfigureLogger(config.LogConfiguration)
	checkErr(config, fs, err)
	exitChannel := elevator.SetupExitChannel()

	if config.Daemon {
		elevator.Daemon(config, exitChannel)
	} else {
		elevator.ListenAndServe(config, exitChannel)
	}
}
