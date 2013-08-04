package elevator

import (
	leveldb "github.com/jmhodges/levigo"
)

type Config struct {
	*CoreConfig
	*StorageEngineConfig
	*LogConfiguration
	*leveldb.Options
}

type CoreConfig struct {
	Daemon      bool   `ini:"daemonize" short:"d" description:"Launches elevator as a daemon"`
	Endpoint    string `ini:"endpoint" short:"e" description:"Endpoint to bind elevator to"`
	Pidfile     string `ini:"pidfile"`
	StorePath   string `ini:"database_store"`
	StoragePath string `ini:"databases_storage_path"`
	DefaultDb   string `ini:"default_db"`
}

type StorageEngineConfig struct {
	Compression     bool `ini:"compression"`       // default: true
	BlockSize       int  `ini:"block_size"`        // default: 4096
	CacheSize       int  `ini:"cache_size"`        // default: 128 * 1048576 (128MB)
	BloomFilterBits int  `ini:"bloom_filter_bits"` // default: 100
	MaxOpenFiles    int  `ini:"max_open_files"`    // default: 150
	VerifyChecksums bool `ini:"verify_checksums"`  // default: false
	WriteBufferSize int  `ini:"write_buffer_size"` // default: 64 * 1048576 (64MB)
}

type LogConfiguration struct {
	LogFile  string `ini:"log_file"`
	LogLevel string `ini:"log_level" short:"l" description:"Sets elevator verbosity"`
}

func NewConfig() *Config {
	storage := NewStorageEngineConfig()
	levelDbOptions := storage.ToLeveldbOptions()
	return &Config{
		NewCoreConfig(),
		storage,
		NewLogConfiguration(),
		levelDbOptions,
	}
}

func NewCoreConfig() *CoreConfig {
	c := &CoreConfig{
		Daemon:      false,
		Endpoint:    DEFAULT_ENDPOINT,
		Pidfile:     "/var/run/elevator.pid",
		StorePath:   "/var/lib/elevator/store.json",
		StoragePath: "/var/lib/elevator",
		DefaultDb:   "default",
	}
	return c
}

func NewLogConfiguration() *LogConfiguration {
	return &LogConfiguration{
		LogFile:  "/var/log/elevator.log",
		LogLevel: "INFO",
	}
}

func NewStorageEngineConfig() *StorageEngineConfig {
	return &StorageEngineConfig{
		Compression:     true,
		BlockSize:       131072,
		CacheSize:       512 * 1048576,
		BloomFilterBits: 100,
		MaxOpenFiles:    150,
		VerifyChecksums: false,
		WriteBufferSize: 64 * 1048576,
	}
}

func (opts StorageEngineConfig) ToLeveldbOptions() *leveldb.Options {
	options := leveldb.NewOptions()

	options.SetCreateIfMissing(true)
	options.SetCompression(leveldb.CompressionOpt(Btoi(opts.Compression)))
	options.SetBlockSize(opts.BlockSize)
	options.SetCache(leveldb.NewLRUCache(opts.CacheSize))
	options.SetFilterPolicy(leveldb.NewBloomFilter(opts.BloomFilterBits))
	options.SetMaxOpenFiles(opts.MaxOpenFiles)
	options.SetParanoidChecks(opts.VerifyChecksums)
	options.SetWriteBufferSize(opts.WriteBufferSize)

	return options
}

func (opts *StorageEngineConfig) UpdateFromConfig(config *Config) {
	opts.Compression = config.Compression
	opts.BlockSize = config.BlockSize
	opts.CacheSize = config.CacheSize
	opts.BloomFilterBits = config.BloomFilterBits
	opts.MaxOpenFiles = config.MaxOpenFiles
	opts.VerifyChecksums = config.VerifyChecksums
	opts.WriteBufferSize = config.WriteBufferSize
}

func ConfFromFile(path string) (*Config, error) {
	conf := NewConfig()
	if err := LoadConfigFromFile(path, conf.CoreConfig, "core"); err != nil {
		return conf, err
	}
	if err := LoadConfigFromFile(path, conf.StorageEngineConfig, "storage_engine"); err != nil {
		return conf, err
	}
	if err := LoadConfigFromFile(path, conf.LogConfiguration, "log"); err != nil {
		return conf, err
	}
	return conf, nil
}
