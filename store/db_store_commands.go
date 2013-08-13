package store

import (
	"errors"
)

var NO_SUCH_DB = errors.New("Database does not exist")

var storeCommands = map[string]func(*DbStore, [][]byte) ([][]byte, error){
	DB_CREATE:  DbCreate,
	DB_DROP:    DbDrop,
	DB_CONNECT: DbConnect,
	DB_MOUNT:   DbMount,
	DB_UMOUNT:  DbUnmount,
	DB_LIST:    DbList,
}

// List enumerates  all the databases
// in DbStore
func (store *DbStore) List() []string {
	db_names := make([]string, len(store.Container))
	i := 0
	for _, db := range store.Container {
		db_names[i] = db.Name
		i++
	}
	return db_names
}

func DbCreate(db_store *DbStore, args [][]byte) ([][]byte, error) { db_name := string(args[0])
	err := db_store.Add(db_name)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

func DbDrop(db_store *DbStore, args [][]byte) ([][]byte, error) {
	db_name := string(args[0])
	err := db_store.Drop(db_name)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

func DbConnect(db_store *DbStore, args [][]byte) ([][]byte, error) {
	db_name := string(args[0])
	db_uid, exists := db_store.NameToUid[db_name]
	if !exists {
		return nil, NoSuchDbError(db_name)
	}
	return ToBytes(db_uid), nil
}

func DbList(db_store *DbStore, args [][]byte) ([][]byte, error) {
	db_names := db_store.List()
	data := make([][]byte, len(db_names))
	for index, db_name := range db_names {
		data[index] = []byte(db_name)
	}
	return data, nil
}

func DbMount(db_store *DbStore, args [][]byte) ([][]byte, error) {
	db_name := string(args[0])
	db_uid, exists := db_store.NameToUid[db_name]
	if !exists {
		return nil, NoSuchDbError(db_name)
	}
	err := db_store.Mount(db_store.Container[db_uid].Name)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

func DbUnmount(db_store *DbStore, args [][]byte) ([][]byte, error) {
	db_name := string(args[0])
	db_uid, exists := db_store.NameToUid[db_name]
	if !exists {
		return nil, NoSuchDbError(db_name)
	}
	err := db_store.Unmount(db_uid)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}
