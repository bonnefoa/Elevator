package main

import (
	"flag"
	"fmt"
	"github.com/oleiade/Elevator/server"
)

func main() {
    cmdConf := server.NewConfig()
	server.SetFlag(cmdConf.ServerConfig)
	confFile := flag.String("c",
		server.DefaultConfigFile,
		"Specifies cmdConf file path")
	flag.Parse()

	fileConfig, err := server.ConfFromFile(*confFile)
    if err != nil {
		fmt.Println(err)
		flag.PrintDefaults()
    }

    if cmdConf.ServerConfig.Daemon != server.DefaultDaemonMode {
        fileConfig.ServerConfig.Daemon = cmdConf.ServerConfig.Daemon
    }
    if cmdConf.ServerConfig.Endpoint != server.DefaultEndpoint {
        fileConfig.ServerConfig.Endpoint = cmdConf.ServerConfig.Endpoint
    }
    if cmdConf.ServerConfig.NumWorkers != server.DefaultWorkers {
        fileConfig.ServerConfig.NumWorkers = cmdConf.ServerConfig.NumWorkers
    }

	exitChannel := server.SetupExitChannel()
	if fileConfig.Daemon {
		server.Daemon(fileConfig, exitChannel)
	} else {
		server.ListenAndServe(fileConfig, exitChannel)
	}
}
