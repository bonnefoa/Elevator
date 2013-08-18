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
    *db
    *StoreConfig
}

func setupEnv(t Tester) *Env {
    env := &Env{Tester:t}
	env.StoreConfig = getTestConf()
	env.DbStore = NewDbStore(env.StoreConfig)
	err := env.Create(TestDb)
	if err != nil {
		env.Fatalf("Error when creating test db %v", err)
	}
	if err != nil {
		env.Fatalf("Error on connection %v", err)
	}
	env.db = env.Container[env.nameToUID[TestDb]]
	if env.db == nil {
		env.Fatalf("No db for %v", TestDb)
	}
	if env.db.status == statusUnmounted {
		env.Fatalf("Db is unmounted %s", TestDb)
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

func fillNKeys(db *db, n int) {
	batchPuts := make([]*BatchPut, n)
	for i := 0; i < n; i ++ {
        key := []byte(fmt.Sprintf("key_%d", i))
        value := []byte(fmt.Sprintf("val_%d", i))
		batchPuts[i] = &BatchPut{key, value, nil}
	}
	db.batch(batchPuts, nil)
}
