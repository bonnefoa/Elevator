package store

import (
	leveldb "github.com/jmhodges/levigo"
	"os"
)

type StoreConfig struct {
	*CoreConfig
	*leveldb.Options
}

type CoreConfig struct {
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

func NewStoreConfig() *StoreConfig {
	core := NewCoreConfig()
	options := NewStorageEngineConfig().ToLeveldbOptions()
	return &StoreConfig { core, options }
}

func NewCoreConfig() *CoreConfig {
	c := &CoreConfig{
		StorePath:   "/var/lib/elevator/store.json",
		StoragePath: "/var/lib/elevator",
		DefaultDb:   "default",
	}
	return c
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

func (c *StoreConfig) CleanConfiguration() {
	os.RemoveAll(c.StoragePath)
	os.RemoveAll(c.StorePath)
}
