package store

import (
)


var storeCommands = map[string]func(*DbStore, [][]byte) ([][]byte, error){
	DbCreate:  Create,
	DbDrop:    Drop,
	DbConnect: Connect,
	DbMount:   Mount,
	DbUmount:  Umount,
	DbList:    List,
}

// List enumerates  all the databases
// in DbStore
func (store *DbStore) List() []string {
	dbNames := make([]string, len(store.Container))
	i := 0
	for _, db := range store.Container {
		dbNames[i] = db.Name
		i++
	}
	return dbNames
}

// Create creates a new database with the given name
// Fail if a database with the same name already exists
func Create(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbName := string(args[0])
	err := dbStore.Add(dbName)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

// Drop drops a data with the given name
// Fail if no database with the given name exists
func Drop(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbName := string(args[0])
	err := dbStore.Drop(dbName)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

// Connect return the matching UID from the given name
func Connect(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbName := string(args[0])
	dbUID, exists := dbStore.nameToUID[dbName]
	if !exists {
		return nil, NoSuchDbError(dbName)
	}
	return ToBytes(dbUID), nil
}

// List lists all available databases from the store
func List(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbNames := dbStore.List()
	data := make([][]byte, len(dbNames))
	for index, dbName := range dbNames {
		data[index] = []byte(dbName)
	}
	return data, nil
}

// Mount open connection to the database with the given name
func Mount(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbName := string(args[0])
	dbUID, exists := dbStore.nameToUID[dbName]
	if !exists {
		return nil, NoSuchDbError(dbName)
	}
	err := dbStore.Mount(dbStore.Container[dbUID].Name)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}

// Umount closes connection to the database with the given name
func Umount(dbStore *DbStore, args [][]byte) ([][]byte, error) {
	dbName := string(args[0])
	dbUID, exists := dbStore.nameToUID[dbName]
	if !exists {
		return nil, NoSuchDbError(dbName)
	}
	err := dbStore.Unmount(dbUID)
	if err != nil {
		return nil, DatabaseError(err)
	}
	return nil, nil
}
