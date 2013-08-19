package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/oleiade/Elevator/server"
	"os"
)

func checkErr(config *server.Config, fs *flag.FlagSet, err error) {
	if err != nil {
		fs.PrintDefaults()
		if config != nil {
			glog.Infof("Current config is %s\n", config)
		}
		glog.Error(err)
	}
}

func main() {
	fs := flag.NewFlagSet("Elevator flag set", flag.ContinueOnError)

	confFile := fs.String("c",
		server.DefaultConfigFile,
		"Specifies config file path")
	err := fs.Parse(os.Args[1:])
	checkErr(nil, fs, err)

	config, err := server.ConfFromFile(*confFile)
	server.SetFlag(fs, config.CoreConfig)
	checkErr(config, fs, err)

	// Reparse to make command line arguments override conf file
	err = fs.Parse(os.Args[:1])
	checkErr(config, fs, err)

	checkErr(config, fs, err)
	exitChannel := server.SetupExitChannel()

	if config.Daemon {
		server.Daemon(config, exitChannel)
	} else {
		server.ListenAndServe(config, exitChannel)
	}
}
