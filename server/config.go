package server

import (
	store "github.com/oleiade/Elevator/store"
)

type Config struct {
	*ServerConfig
	*store.StoreConfig
	*LogConfiguration
}

type ServerConfig struct {
	Daemon   bool   `ini:"daemonize" short:"d" description:"Launches elevator as a daemon"`
	Endpoint string `ini:"endpoint" short:"e" description:"Endpoint to bind elevator to"`
	Pidfile  string `ini:"pidfile"`
}

type LogConfiguration struct {
	LogFile  string `ini:"log_file"`
	LogLevel string `ini:"log_level" short:"l" description:"Sets elevator verbosity"`
}

func NewConfig() *Config {
	storeConfig := store.NewStoreConfig()
	serverConfig := NewServerConfig()
	logConfiguration := NewLogConfiguration()
	return &Config{serverConfig, storeConfig, logConfiguration}
}

func NewServerConfig() *ServerConfig {
	c := &ServerConfig{
		Daemon:   false,
		Endpoint: DEFAULT_ENDPOINT,
		Pidfile:  "/var/run/elevator.pid",
	}
	return c
}

func NewLogConfiguration() *LogConfiguration {
	return &LogConfiguration{
		LogFile:  "/var/log/elevator.log",
		LogLevel: "INFO",
	}
}

func ConfFromFile(path string) (*Config, error) {
	conf := NewConfig()
	storageConfig := store.NewStorageEngineConfig()
	if err := LoadConfigFromFile(path, conf.ServerConfig, "core"); err != nil {
		return conf, err
	}
	if err := LoadConfigFromFile(path, &storageConfig, "storage_engine"); err != nil {
		return conf, err
	}
	conf.StoreConfig.Options = storageConfig.ToLeveldbOptions()
	if err := LoadConfigFromFile(path, conf.LogConfiguration, "log"); err != nil {
		return conf, err
	}
	return conf, nil
}
