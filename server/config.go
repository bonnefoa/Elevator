package server

import (
	store "github.com/oleiade/Elevator/store"
)

// Config keeps configuration of elevator
type Config struct {
	*serverConfig
	*store.StoreConfig
}

type serverConfig struct {
	Daemon   bool   `ini:"daemonize" short:"d" description:"Launches elevator as a daemon"`
	Endpoint string `ini:"endpoint" short:"e" description:"Endpoint to bind elevator to"`
	Pidfile  string `ini:"pidfile"`
	NumWorkers  int `ini:"numworkers" short:"n" description:"The number of goroutine workers to launch"`
}

func newConfig() *Config {
	storeConfig := store.NewStoreConfig()
	serverConfig := newServerConfig()
	return &Config{serverConfig, storeConfig}
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

// ConfFromFile reads a server config from the given file
func ConfFromFile(path string) (*Config, error) {
	conf := newConfig()
	storageConfig := store.NewStorageEngineConfig()
	if err := loadConfigFromFile(path, conf.serverConfig, "core"); err != nil {
		return conf, err
	}
	if err := loadConfigFromFile(path, &storageConfig, "storage_engine"); err != nil {
		return conf, err
	}
	conf.StoreConfig.Options = storageConfig.ToLeveldbOptions()
	return conf, nil
}
