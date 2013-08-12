package store

import (
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"io/ioutil"
	"os"
)

const TestDb = "test_db"

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func getTestConf() *StoreConfig {
	storePath, _ := ioutil.TempFile("/tmp", "elevator_store")
	storagePath, _ := ioutil.TempDir("/tmp", "elevator_path")

	core := &CoreConfig{
		StorePath:   storePath.Name(),
		StoragePath: storagePath,
		DefaultDb:   "default",
	}
	storage := NewStorageEngineConfig()
	options := storage.ToLeveldbOptions()
	config := &StoreConfig{
		core,
		options,
	}
	l4g.AddFilter("stdout", l4g.INFO, l4g.NewConsoleLogWriter())
	return config
}

func cleanConfTemp(c *StoreConfig) {
	os.RemoveAll(c.StoragePath)
	os.RemoveAll(c.StorePath)
}

func fillNKeys(db *Db, n int) {
	req := make([]string, n*3)
	for i := 0; i < n*3; i += 3 {
		req[i] = SIGNAL_BATCH_PUT
		req[i+1] = fmt.Sprintf("key_%i", i)
		req[i+2] = fmt.Sprintf("val_%i", i)
	}
	Batch(db, toBytes(req...))
}

func TemplateDbTest(t Tester, f func(*DbStore, *Db)) {
	c := getTestConf()
	defer cleanConfTemp(c)
	db_store := NewDbStore(c)
	err := db_store.Add(TestDb)
	if err != nil {
		t.Fatalf("Error when creating test db %v", err)
	}
	res, err := DbConnect(db_store, toBytes(TestDb))
	if err != nil {
		t.Fatalf("Error on connection %v", err)
	}
	dbUid := string(res[0])
	db := db_store.Container[dbUid]
	if db == nil {
		t.Fatalf("No db for uid %v", dbUid)
	}
	if db.status == DB_STATUS_UNMOUNTED {
		t.Fatalf("Db is unmounted %s", dbUid)
	}

	defer db_store.UnmountAll()
	if err != nil {
		t.Fatalf("Error when creating test db %v", err)
	}
	f(db_store, db)
}

