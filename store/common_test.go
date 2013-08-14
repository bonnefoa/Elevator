package store

import (
	"fmt"
	"io/ioutil"
)

const TestDb = "test_db"

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func getTestConf() *StoreConfig {
	storeConfig := NewStoreConfig()
	storePath, _ := ioutil.TempFile("/tmp", "elevator_store")
	storagePath, _ := ioutil.TempDir("/tmp", "elevator_path")
	storeConfig.CoreConfig.StorePath = storePath.Name()
	storeConfig.CoreConfig.StoragePath = storagePath
	return storeConfig
}

type Env struct {
	Tester
    *DbStore
    *Db
    *StoreConfig
}

func setupEnv(t Tester) *Env {
    env := &Env{Tester:t}
	env.StoreConfig = getTestConf()
	env.DbStore = NewDbStore(env.StoreConfig)
	err := env.Add(TestDb)
	if err != nil {
		env.Fatalf("Error when creating test db %v", err)
	}
	res, err := DbConnect(env.DbStore, ToBytes(TestDb))
	if err != nil {
		env.Fatalf("Error on connection %v", err)
	}
	dbUid := string(res[0])
	env.Db = env.Container[dbUid]
	if env.Db == nil {
		env.Fatalf("No db for uid %v", dbUid)
	}
	if env.Db.status == DB_STATUS_UNMOUNTED {
		env.Fatalf("Db is unmounted %s", dbUid)
	}
	if err != nil {
		env.Fatalf("Error when creating test db %v", err)
	}
    return env
}

func (env *Env) destroy() {
	env.UnmountAll()
    env.CleanConfiguration()
}

func fillNKeys(db *Db, n int) {
	req := make([]string, n*3)
	for i := 0; i < n*3; i += 3 {
		req[i] = SIGNAL_BATCH_PUT
		req[i+1] = fmt.Sprintf("key_%i", i)
		req[i+2] = fmt.Sprintf("val_%i", i)
	}
	Batch(db, ToBytes(req...))
}
