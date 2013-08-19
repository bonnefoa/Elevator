package server

import (
	store "github.com/oleiade/Elevator/store"
)

type config struct {
	*serverConfig
	*store.StoreConfig
	*logConfiguration
}

type serverConfig struct {
	Daemon   bool   `ini:"daemonize" short:"d" description:"Launches elevator as a daemon"`
	Endpoint string `ini:"endpoint" short:"e" description:"Endpoint to bind elevator to"`
	Pidfile  string `ini:"pidfile"`
	NumWorkers  int `ini:"numworkers" short:"n" description:"The number of goroutine workers to launch"`
}

type logConfiguration struct {
	LogFile  string `ini:"log_file"`
	LogLevel string `ini:"log_level" short:"l" description:"Sets elevator verbosity"`
}

func newConfig() *config {
	storeConfig := store.NewStoreConfig()
	serverConfig := newServerConfig()

	logConfiguration := newLogConfiguration()
	return &config{serverConfig, storeConfig, logConfiguration}
}

func newServerConfig() *serverConfig {
	c := &serverConfig{
		Daemon:   false,
		Endpoint: DefaultEndpoint,
		Pidfile:  "/var/run/elevator.pid",
        NumWorkers: 5,
	}
	return c
}

func newLogConfiguration() *logConfiguration {
	return &logConfiguration{
		LogFile:  "/var/log/elevator.log",
		LogLevel: "INFO",
	}
}

func confFromFile(path string) (*config, error) {
	conf := newConfig()
	storageConfig := store.NewStorageEngineConfig()
	if err := loadConfigFromFile(path, conf.serverConfig, "core"); err != nil {
		return conf, err
	}
	if err := loadConfigFromFile(path, &storageConfig, "storage_engine"); err != nil {
		return conf, err
	}
	conf.StoreConfig.Options = storageConfig.ToLeveldbOptions()
	if err := loadConfigFromFile(path, conf.logConfiguration, "log"); err != nil {
		return conf, err
	}
	return conf, nil
}
