package store

import (
	"bytes"
	leveldb "github.com/jmhodges/levigo"
	"strconv"
)

var databaseComands = map[string]func(*db, [][]byte) ([][]byte, error){
	DbGet:    get,
	DbMget:   mGet,
	DbPut:    put,
	DbDelete: dbDelete,
	DbRange:  dbRange,
	DbSlice:  slice,
	DbBatch:  batch,
}

func get(db *db, args [][]byte) ([][]byte, error) {
	key := args[0]
	readOptions := leveldb.NewReadOptions()
	value, err := db.connector.Get(readOptions, key)
	if value == nil {
		return nil, KeyError(key)
	} else if err != nil {
		return nil, DatabaseError(err)
	}
	return [][]byte{value}, nil
}

func put(db *db, args [][]byte) ([][]byte, error) {
	key := args[0]
	value := args[1]
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Put(writeOptions, key, value)
	if err != nil {
		return nil, ValueError(err)
	}
	return nil, nil
}

func dbDelete(db *db, args[][]byte) ([][]byte, error) {
	key := args[0]
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Delete(writeOptions, key)
	if err != nil {
		return nil, KeyError(key)
	}
	return nil, nil
}

func mGet(db *db, args [][]byte) ([][]byte, error) {
	data := make([][]byte, len(args))
	readOptions := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)
	for i, key := range args {
		value, _ := db.connector.Get(readOptions, key)
		data[i] = value
	}
	db.connector.ReleaseSnapshot(snapshot)
	return data, nil
}

func dbRange(db *db, args[][]byte) ([][]byte, error) {
	var data [][]byte
	start := args[0]
	end := args[1]

	readOptions := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)

	it := db.connector.NewIterator(readOptions)
	defer it.Close()
	it.Seek(start)

	for ; it.Valid(); it.Next() {
		if bytes.Compare(it.Key(), end) >= 1 {
			break
		}
		data = append(data, it.Key(), it.Value())
	}
	db.connector.ReleaseSnapshot(snapshot)
	return data, nil
}

func slice(db *db, args [][]byte) ([][]byte, error) {
	var data [][]byte
	start := args[0]
	limit, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, RequestError(err)
	}
	readOptions := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)
	it := db.connector.NewIterator(readOptions)
	defer it.Close()
	it.Seek([]byte(start))
	i := 0
	for ; it.Valid(); it.Next() {
		if i >= limit {
			break
		}
		data = append(data, it.Key(), it.Value())
		i++
	}
	db.connector.ReleaseSnapshot(snapshot)
	return data, nil
}

func batch(db *db, args [][]byte) ([][]byte, error) {
	batch := leveldb.NewWriteBatch()
	operations, err := batchOperationsFromRequestArgs(args)
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		operation.executeBatch(batch)
	}
	wo := leveldb.NewWriteOptions()
	err = db.connector.Write(wo, batch)
	if err != nil {
		return nil, ValueError(err)
	}
	return nil, nil
}
