package server

import (
	"github.com/golang/glog"
	store "github.com/oleiade/Elevator/store"
)

// Config keeps configuration of elevator
type Config struct {
	*ServerConfig
	*store.StoreConfig
}

// ServerConfig stores server specific configuration
type ServerConfig struct {
	Daemon     bool   `ini:"daemonize" short:"d" description:"Launches elevator as a daemon"`
	Endpoint   string `ini:"endpoint" short:"e" description:"Endpoint to bind elevator to"`
	Pidfile    string `ini:"pidfile"`
	NumWorkers int    `ini:"numworkers" short:"n" description:"The number of goroutine workers to launch"`
}

// NewConfig creates a new instance of server config with default parameters
func NewConfig() *Config {
	storeConfig := store.NewStoreConfig()
	serverConfig := newServerConfig()
	return &Config{serverConfig, storeConfig}
}

func newServerConfig() *ServerConfig {
	c := &ServerConfig{
		Daemon:     DefaultDaemonMode,
		Endpoint:   DefaultEndpoint,
		Pidfile:    DefaultPidfile,
		NumWorkers: DefaultWorkers,
	}
	return c
}

// ConfFromFile reads a server config from the given file
func ConfFromFile(path string) (*Config, error) {
	glog.Infof("Loading configuration file %s", path)
	conf := NewConfig()
	storageConfig := store.NewStorageEngineConfig()
	if err := loadConfigFromFile(path, conf.ServerConfig, "server"); err != nil {
		return conf, err
	}
	if err := loadConfigFromFile(path, conf.CoreConfig, "core"); err != nil {
		return conf, err
	}
	if err := loadConfigFromFile(path, storageConfig, "storage_engine"); err != nil {
		return conf, err
	}
	conf.StoreConfig.Options = storageConfig.ToLeveldbOptions()
	return conf, nil
}
