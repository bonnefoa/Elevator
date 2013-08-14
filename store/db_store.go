package store

import (
	"encoding/json"
	"errors"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DbStore struct {
	*StoreConfig
	Container map[string]*Db
	NameToUid map[string]string
}

// DbStore constructor
func NewDbStore(config *StoreConfig) *DbStore {
	return &DbStore{
		config,
		make(map[string]*Db),
		make(map[string]string),
	}
}

func InitializeDbStore(storeConfig *StoreConfig) (*DbStore, error) {
	dbStore := NewDbStore(storeConfig)
	err := dbStore.Load()
	if err != nil {
		err = dbStore.Add(storeConfig.DefaultDb)
	}
	return dbStore, err
}

func (store *DbStore) updateNameToUidIndex() {
	for _, db := range store.Container {
		if _, present := store.NameToUid[db.Name]; present == false {
			store.NameToUid[db.Name] = db.Uid
		}
	}
}

// ReadFromFile syncs the content of the store
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
	store.updateNameToUidIndex()
	return nil
}

// WriteToFile syncs the content of the DbStore
// to the store description file
func (store *DbStore) WriteToFile() (err error) {
	var data []byte
	// Check the directory hosting the store exists
	store_base_path := filepath.Dir(store.StorePath)
	_, err = os.Stat(store_base_path)
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

// Mount sets the database status to DB_STATUS_MOUNTED
// and instantiates the according leveldb connector
func (store *DbStore) Mount(db_uid string) (err error) {
	if db, present := store.Container[db_uid]; present {
		err = db.Mount(store.StoreConfig.Options)
		if err != nil {
			return err
		}
	} else {
		error := fmt.Errorf("Database with uid %s does not exist", db_uid)
		l4g.Error(error)
		return error
	}
	return nil
}

// Unmount sets the database status to DB_STATUS_UNMOUNTED
// and deletes the according leveldb connector
func (store *DbStore) Unmount(db_uid string) (err error) {
	if db, present := store.Container[db_uid]; present {
		err = db.Unmount()
		if err != nil {
			return err
		}
	} else {
		error := fmt.Errorf("Database with uid %s does not exist", db_uid)
		l4g.Error(error)
		return error
	}
	return nil
}

// Add a db to the DbStore and syncs it
// to the store file
func (store *DbStore) Add(db_name string) (err error) {
	if _, present := store.NameToUid[db_name]; present {
		return errors.New("Database already exists")
	} else {
		var db_path string

		if IsFilePath(db_name) {
			if !filepath.IsAbs(db_name) {
				error := errors.New("Creating database from relative path not allowed")
				l4g.Error(error)
				return error
			}
			db_path = db_name
			// Check base db path exists
			dir := filepath.Dir(db_name)
			exists, err := DirExists(dir)
			if err != nil {
				l4g.Error(err)
				return err
			} else if !exists {
				error := fmt.Errorf("%s does not exist", dir)
				l4g.Error(error)
				return error
			}
		} else {
			db_path = filepath.Join(store.StoragePath, db_name)
		}
		db := NewDb(db_name, db_path)
		store.Container[db.Uid] = db
		store.updateNameToUidIndex()
		err = store.WriteToFile()
		if err != nil {
			l4g.Error(err)
			return err
		}
		err = db.Mount(store.StoreConfig.Options)
		if err != nil {
			l4g.Error(err)
			return err
		}
	}
	l4g.Debug(func() string {
		return fmt.Sprintf("Database %s added to store", db_name)
	})
	return nil
}

// Drop removes a database from DbStore, and syncs it
// to store file
func (store *DbStore) Drop(db_name string) (err error) {
	if db_uid, present := store.NameToUid[db_name]; present {
		db := store.Container[db_uid]
		db_path := db.Path

		store.Unmount(db_uid)
		delete(store.Container, db_uid)
		delete(store.NameToUid, db_name)
		store.WriteToFile()

		err = os.RemoveAll(db_path)
		if err != nil {
			l4g.Error(err)
			return err
		}
	} else {
		error := fmt.Errorf("Database %s does not exist", db_name)
		l4g.Error(error)
		return error
	}
	l4g.Debug(func() string {
		return fmt.Sprintf("Database %s dropped from store", db_name)
	})
	return nil
}

// Status returns a database status defined by constants
// DB_STATUS_MOUNTED and DB_STATUS_UNMOUNTED
func (store *DbStore) Status(db_name string) (DbMountedStatus, error) {
	if db_uid, present := store.NameToUid[db_name]; present {
		db := store.Container[db_uid]
		return db.status, nil
	}
	return -1, errors.New("Database does not exist")
}

// Exists checks if a database present in DbStore
// exists on disk.
func (store *DbStore) Exists(db_name string) (bool, error) {
	if db_uid, present := store.NameToUid[db_name]; present {
		db := store.Container[db_uid]
		exists, err := DirExists(db.Path)
		if err != nil {
			return false, err
		}
		if exists == true {
			return exists, nil
		} else {
			// store.drop(db_name)
			fmt.Println("Dropping")
		}
	}
	return false, nil
}

func (store *DbStore) UnmountAll() {
	l4g.Debug("Closing dbstore")
	for _, db := range store.Container {
		if db.status == DB_STATUS_MOUNTED {
			db.Unmount()
		}
	}
}

// Redirect and execute given request as a db operation or a store operation
func (dbStore *DbStore) HandleRequest(request *Request) ([][]byte, error) {
	switch request.TypeCommand {
	case STORE_COMMAND:
		res, err := storeCommands[request.Command](dbStore, request.Args)
		return res, err
	case DB_COMMAND:
		db, foundDb := dbStore.Container[request.DbUid]
		if !foundDb {
			return nil, NoSuchDbuidError(request.DbUid)
		}
		if db.status == DB_STATUS_UNMOUNTED {
			err := db.Mount(dbStore.Options)
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
