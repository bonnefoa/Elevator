package store

import (
	"bytes"
	leveldb "github.com/jmhodges/levigo"
	"strconv"
)

var databaseComands = map[string]func(*Db, [][]byte) ([][]byte, error){
	DB_GET:    Get,
	DB_MGET:   MGet,
	DB_PUT:    Put,
	DB_DELETE: Delete,
	DB_RANGE:  Range,
	DB_SLICE:  Slice,
	DB_BATCH:  Batch,
}

func Get(db *Db, args [][]byte) ([][]byte, error) {
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

func Put(db *Db, args [][]byte) ([][]byte, error) {
	key := args[0]
	value := args[1]
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Put(writeOptions, key, value)
	if err != nil {
		return nil, ValueError(err)
	}
	return nil, nil
}

func Delete(db *Db, args[][]byte) ([][]byte, error) {
	key := args[0]
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Delete(writeOptions, key)
	if err != nil {
		return nil, KeyError(key)
	}
	return nil, nil
}

func MGet(db *Db, args [][]byte) ([][]byte, error) {
	var data [][]byte = make([][]byte, len(args))
	read_options := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	read_options.SetSnapshot(snapshot)
	for i, key := range args {
		value, _ := db.connector.Get(read_options, key)
		data[i] = value
	}
	db.connector.ReleaseSnapshot(snapshot)
	return data, nil
}

func Range(db *Db, args[][]byte) ([][]byte, error) {
	var data [][]byte
	start := args[0]
	end := args[1]

	read_options := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	read_options.SetSnapshot(snapshot)

	it := db.connector.NewIterator(read_options)
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

func Slice(db *Db, args [][]byte) ([][]byte, error) {
	var data [][]byte
	start := args[0]
	limit, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return nil, RequestError(err)
	}
	read_options := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	read_options.SetSnapshot(snapshot)
	it := db.connector.NewIterator(read_options)
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

func Batch(db *Db, args [][]byte) ([][]byte, error) {
	var batch *leveldb.WriteBatch = leveldb.NewWriteBatch()
	operations, err := BatchOperationsFromRequestArgs(args)
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		operation.ExecuteBatch(batch)
	}
	wo := leveldb.NewWriteOptions()
	err = db.connector.Write(wo, batch)
	if err != nil {
		return nil, ValueError(err)
	}
	return nil, nil
}
