package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/golang/glog"
)

// DbStore allow to create / drop and mount datastores
type DbStore struct {
	*StoreConfig
	Container map[string]*db
	nameToUID map[string]string
}

// NewDbStore build a dbstore from the given config
func NewDbStore(config *StoreConfig) *DbStore {
	return &DbStore{
		config,
		make(map[string]*db),
		make(map[string]string),
	}
}

// InitializeDbStore create and initialize dbStore from given configuration
func InitializeDbStore(storeConfig *StoreConfig) (*DbStore, error) {
	dbStore := NewDbStore(storeConfig)
	err := dbStore.Load()
	if err != nil {
		err = dbStore.Create(storeConfig.DefaultDb)
	}
	return dbStore, err
}

func (store *DbStore) udpateNameToUIDIndex() {
	for _, db := range store.Container {
		if _, present := store.nameToUID[db.Name]; !present {
			store.nameToUID[db.Name] = db.UID
		}
	}
}

// Load syncs the content of the store
// description file to the DbStore
func (store *DbStore) Load() error {
	data, err := ioutil.ReadFile(store.StorePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &store.Container)
	if err != nil {
		return err
	}
	store.udpateNameToUIDIndex()
	return nil
}

// List enumerates  all the databases
// in DbStore
func (store *DbStore) List() [][]byte {
       dbNames := make([][]byte, len(store.Container))
       i := 0
       for _, db := range store.Container {
               dbNames[i] = []byte(db.Name)
               i++
       }
       return dbNames
}

// WriteToFile syncs the content of the DbStore
// to the store description file
func (store *DbStore) WriteToFile() error {
	var data []byte
	// Check the directory hosting the store exists
	storeBasePath := filepath.Dir(store.StorePath)
    _, err := os.Stat(storeBasePath)
	if os.IsNotExist(err) {
		return err
	}
	data, err = json.Marshal(store.Container)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(store.StorePath, data, 0777)
	return err
}

// Mount sets the database status to statusMounted
// and instantiates the according leveldb connector
func (store *DbStore) Mount(dbUID string) error {
    db, present := store.Container[dbUID]
	if !present {
		return NoSuchDbError(dbUID)
    }
    if db.status == statusMounted {
        return nil
    }
    err := db.Mount(store.StoreConfig.Options)
    if err != nil {
        return DatabaseError(err)
    }
	return nil
}

// Umount sets the database status to statusUnmounted
// and deletes the according leveldb connector
func (store *DbStore) Umount(dbUID string) error {
    db, present := store.Container[dbUID]
	if !present {
        return NoSuchDbError(dbUID)
    }
    err := db.Umount()
    if err != nil {
        return DatabaseError(err)
    }
	return nil
}

func (store *DbStore) getPathFromDbname(dbName string) (dbPath string, err error) {
	if !isFilePath(dbName) {
        dbPath = filepath.Join(store.StoragePath, dbName)
        return
    }
    if !filepath.IsAbs(dbName) {
        err = RelativePathError(dbName)
        return
    }
    // Check base db path exists
    dir := filepath.Dir(dbName)
    exists, err := dirExists(dir)
    if err != nil {
        return
    }
    if !exists {
        err = NoSuchPathError(dbName)
        return
    }
    return dbName, nil
}

// Create a db to the DbStore and syncs it
// to the store file
func (store *DbStore) Create(dbName string) error {
	if _, present := store.nameToUID[dbName]; present {
		return DatabaseExistsError(dbName)
	}
    dbPath, err := store.getPathFromDbname(dbName)
    if err != nil {
        return err
    }
	db := newDb(dbName, dbPath)
	store.Container[db.UID] = db
	store.udpateNameToUIDIndex()
	err = store.WriteToFile()
	if err != nil {
		return err
	}
	err = db.Mount(store.StoreConfig.Options)
	if err != nil {
		return err
	}
	if glog.V(2) {
        glog.Info(fmt.Sprintf("Database %s added to store", dbName))
	}
	return nil
}

// Drop removes a database from DbStore, and syncs it
// to store file
func (store *DbStore) Drop(dbName string) (err error) {
	dbUID, present := store.nameToUID[dbName]
    if !present {
		return NoSuchDbError(dbName)
    }
    db := store.Container[dbUID]
    store.Umount(dbUID)
    delete(store.Container, dbUID)
    delete(store.nameToUID, dbName)
    store.WriteToFile()
    err = os.RemoveAll(db.Path)
    if err == nil && glog.V(2) {
        glog.Info("Database %s dropped from store", dbName)
    }
    return err
}

// Status returns a database status defined by constants
// statusMounted and statusUnmounted
func (store *DbStore) Status(dbName string) (DbMountedStatus, error) {
	if dbUID, present := store.nameToUID[dbName]; present {
		db := store.Container[dbUID]
		return db.status, nil
	}
	return -1, NoSuchDbError(dbName)
}

// Exists checks if a database present in DbStore
// exists on disk.
func (store *DbStore) Exists(dbName string) (bool, error) {
	dbUID, present := store.nameToUID[dbName]
    if !present {
        return false, NoSuchDbError(dbName)
    }
    db := store.Container[dbUID]
    exists, err := dirExists(db.Path)
    return exists, err
}

// UnmountAll Umount all db in datastore
func (store *DbStore) UnmountAll() {
    if glog.V(1) {
        glog.Info("Closing dbstore")
    }
	for _, db := range store.Container {
		if db.status == statusMounted {
			db.Umount()
		}
	}
}

// HandleStoreRequest process store incoming request
func (store *DbStore) HandleStoreRequest(r *StoreRequest) (res [][]byte, err error) {
    err = CheckStoreRequest(r)
    if err != nil {
        return
    }
	switch *r.Command {
	case StoreRequest_CREATE:
        err = store.Create(*r.DbName)
    case StoreRequest_DROP:
        err = store.Drop(*r.DbName)
    case StoreRequest_LIST:
        res = store.List()
	}
    return
}

// HandleDbRequest fetch db from dbname and send dbrequest to the db
func (store *DbStore) HandleDbRequest(r *DbRequest) ([][]byte, error) {
    err := CheckDbRequest(r)
    if err != nil {
        return nil, err
    }
    dbUID, foundDb := store.nameToUID[*r.DbName]
    if !foundDb {
        return nil, NoSuchDbError(*r.DbName)
    }
    db, foundDb := store.Container[dbUID]
    if !foundDb {
        return nil, NoSuchDbUIDError(dbUID)
    }
    err = db.Mount(store.Options)
    if err != nil {
        return nil, err
    }
    res, err := db.processRequest(r)
    return res, err
}

// HandleRequest fetch db from dbname and send dbrequest to the db
func (store *DbStore) HandleRequest(r *Request) (res [][]byte, err error) {
    if glog.V(4) {
        glog.Infof("Handling request %s", r)
    }
    switch *r.Command {
    case Request_DB:
        res, err = store.HandleDbRequest(r.DbRequest)
    case Request_STORE:
        res, err = store.HandleStoreRequest(r.StoreRequest)
    }
    return
}
