package store

import (
	"encoding/json"
	"errors"
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
		err = dbStore.Add(storeConfig.DefaultDb)
	}
	return dbStore, err
}

func (store *DbStore) udpateNameToUIDIndex() {
	for _, db := range store.Container {
		if _, present := store.nameToUID[db.Name]; present == false {
			store.nameToUID[db.Name] = db.UID
		}
	}
}

// Load syncs the content of the store
// description file to the DbStore
func (store *DbStore) Load() (err error) {
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

// WriteToFile syncs the content of the DbStore
// to the store description file
func (store *DbStore) WriteToFile() (err error) {
	var data []byte
	// Check the directory hosting the store exists
	storeBasePath := filepath.Dir(store.StorePath)
	_, err = os.Stat(storeBasePath)
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
func (store *DbStore) Mount(dbUID string) (err error) {
	if db, present := store.Container[dbUID]; present {
		err = db.Mount(store.StoreConfig.Options)
		if err != nil {
			return err
		}
	} else {
		error := fmt.Errorf("Database with uid %s does not exist", dbUID)
		glog.Error(error)
		return error
	}
	return nil
}

// Unmount sets the database status to statusUnmounted
// and deletes the according leveldb connector
func (store *DbStore) Unmount(dbUID string) (err error) {
	if db, present := store.Container[dbUID]; present {
		err = db.Unmount()
		if err != nil {
			return err
		}
	} else {
		error := fmt.Errorf("Database with uid %s does not exist", dbUID)
		glog.Error(error)
		return error
	}
	return nil
}

// Add a db to the DbStore and syncs it
// to the store file
func (store *DbStore) Add(dbName string) (err error) {
	if _, present := store.nameToUID[dbName]; present {
		return errors.New("Database already exists")
	}
	var dbPath string
	if isFilePath(dbName) {
		if !filepath.IsAbs(dbName) {
			error := errors.New("Creating database from relative path not allowed")
			glog.Error(error)
			return error
		}
		dbPath = dbName
		// Check base db path exists
		dir := filepath.Dir(dbName)
		exists, err := dirExists(dir)
		if err != nil {
			glog.Error(err)
			return err
		} else if !exists {
			error := fmt.Errorf("%s does not exist", dir)
			glog.Error(error)
			return error
		}
	} else {
		dbPath = filepath.Join(store.StoragePath, dbName)
	}
	db := newDb(dbName, dbPath)
	store.Container[db.UID] = db
	store.udpateNameToUIDIndex()
	err = store.WriteToFile()
	if err != nil {
		glog.Error(err)
		return err
	}
	err = db.Mount(store.StoreConfig.Options)
	if err != nil {
		glog.Error(err)
		return err
	}
	glog.Info(func() string {
		return fmt.Sprintf("Database %s added to store", dbName)
	})
	return nil
}

// Drop removes a database from DbStore, and syncs it
// to store file
func (store *DbStore) Drop(dbName string) (err error) {
	if dbUID, present := store.nameToUID[dbName]; present {
		db := store.Container[dbUID]
		dbPath := db.Path

		store.Unmount(dbUID)
		delete(store.Container, dbUID)
		delete(store.nameToUID, dbName)
		store.WriteToFile()

		err = os.RemoveAll(dbPath)
		if err != nil {
			glog.Error(err)
			return err
		}
	} else {
		error := fmt.Errorf("Database %s does not exist", dbName)
		glog.Error(error)
		return error
	}
	glog.Info(func() string {
		return fmt.Sprintf("Database %s dropped from store", dbName)
	})
	return nil
}

// Status returns a database status defined by constants
// statusMounted and statusUnmounted
func (store *DbStore) Status(dbName string) (DbMountedStatus, error) {
	if dbUID, present := store.nameToUID[dbName]; present {
		db := store.Container[dbUID]
		return db.status, nil
	}
	return -1, errors.New("Database does not exist")
}

// Exists checks if a database present in DbStore
// exists on disk.
func (store *DbStore) Exists(dbName string) (bool, error) {
	if dbUID, present := store.nameToUID[dbName]; present {
		db := store.Container[dbUID]
		exists, err := dirExists(db.Path)
		if err != nil {
			return false, err
		}
		if exists == true {
			return exists, nil
		}
		// store.drop(dbName)
		fmt.Println("Dropping")
	}
	return false, nil
}

// UnmountAll unmount all db in datastore
func (store *DbStore) UnmountAll() {
	glog.Info("Closing dbstore")
	for _, db := range store.Container {
		if db.status == statusMounted {
			db.Unmount()
		}
	}
}

// HandleRequest redirect and execute given request as a db operation or a store operation
func (store *DbStore) HandleRequest(request *Request) ([][]byte, error) {
	switch request.requestType {
	case typeStore:
		res, err := storeCommands[request.Command](store, request.Args)
		return res, err
	case typeDb:
		db, foundDb := store.Container[request.dbUID]
		if !foundDb {
			return nil, NoSuchDbUIDError(request.dbUID)
		}
		if db.status == statusUnmounted {
			err := db.Mount(store.Options)
			if err != nil {
				return nil, err
			}
		}
		db.requestChan <- request
        result := <-db.responseChan
		return result.Data, result.Err
	}
	return nil, UnknownCommand(request.Command)
}
