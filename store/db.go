package store

import (
	"bytes"
	uuid "code.google.com/p/go-uuid/uuid"
	leveldb "github.com/jmhodges/levigo"
	"github.com/golang/glog"
)

type dbResult struct {
	Data [][]byte
	Err error
}

type db struct {
	Name         string          `json:"name"`
	UID          string          `json:"uid"`
	Path         string          `json:"path"`
	status       DbMountedStatus `json:"-"`
	connector    *leveldb.DB     `json:"-"`
	requestChan  chan *DbRequest   `json:"-"`
	responseChan chan *dbResult  `json:"-"`
}

func newDb(dbName string, path string) *db {
	return &db{
		Name:         dbName,
		Path:         path,
		UID:          uuid.New(),
		status:       statusUnmounted,
		requestChan:  nil,
		responseChan: nil,
	}
}

// processRequest executes the received request command, and returns
// the resulting response.
func (db *db) processRequest(r *DbRequest) (res [][]byte, err error) {
    switch *r.Command{
    case DbRequest_GET:
        var val []byte
        val, err = db.get(r.Get.Key)
        if val != nil {
            res = [][]byte{val}
        }
    case DbRequest_PUT:
        err = db.put(r.Put.Key, r.Put.Value)
    case DbRequest_MGET:
        res, err = db.mget(r.Mget.Keys)
    case DbRequest_RANGE:
        res, err = db.dbRange(r.Range.Start, r.Range.End)
    case DbRequest_SLICE:
        res, err = db.slice(r.Slice.Start, int(*r.Slice.Limit))
    case DbRequest_DELETE:
        err = db.dbDelete(r.Delete.Key)
    case DbRequest_BATCH:
        err = db.batch(r.Batch.Puts, r.Batch.Deletes)
    }
    return
}

// StartRoutine listens on the Db channel awaiting
// for incoming requests to execute. Willingly
// blocking on each Exec call received through the
// channel in order to protect requests.
func (db *db) StartRoutine() {
	for r := range db.requestChan {
		res, err := db.processRequest(r)
		if err != nil {
			glog.Info(err)
		}
		result := &dbResult{
			Data:res,
			Err:err,
		}
		db.responseChan <- result
	}
}

// Mount sets the database status to statusMounted
// and instantiates the according leveldb connector
func (db *db) Mount(options *leveldb.Options) (err error) {
	if db.status == statusMounted {
		return DbAlreadyMounted(db.Name)
    }
    db.connector, err = leveldb.Open(db.Path, options)
    if err != nil {
        return DatabaseError(err)
    }
    db.status = statusMounted
    db.requestChan = make(chan *DbRequest, 100)
    db.responseChan = make(chan *dbResult, 100)
    go db.StartRoutine()
    if glog.V(2) {
        glog.Info("Database %s mounted", db.Name)
    }
    return nil
}

// Umount sets the database status to statusUnmounted
// and deletes the according leveldb connector
func (db *db) Umount() (err error) {
	if db.status == statusUnmounted {
		return DbAlreadyUnmounted(db.Name)
    }
    db.connector.Close()
    close(db.requestChan)
    close(db.responseChan)
    db.status = statusUnmounted
	if glog.V(2) {
        glog.Infoln("Db: Database %s unmounted", db.Name)
    }
	return nil
}


func (db *db) get(key []byte) ([]byte, error) {
	readOptions := leveldb.NewReadOptions()
	value, err := db.connector.Get(readOptions, key)
	if value == nil {
		return nil, KeyError(key)
	}
    if err != nil {
		return nil, DatabaseError(err)
	}
    if glog.V(8) {
        glog.Infof("%q | %q fetched from %q", key, value, db.Name)
    }
	return value, nil
}

func (db *db) put(key []byte, value []byte) error {
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Put(writeOptions, key, value)
	if err != nil {
		return ValueError(err)
	}
    if glog.V(8) {
        glog.Infof("%q | %q added to db %q", key, value, db.Name)
    }
	return nil
}

func (db *db) dbDelete(key []byte)  error {
	writeOptions := leveldb.NewWriteOptions()
	err := db.connector.Delete(writeOptions, key)
	if err != nil {
		return KeyError(key)
	}
	return nil
}

func (db *db) mget(keys [][]byte) ([][]byte, error) {
	data := make([][]byte, len(keys))
	readOptions := leveldb.NewReadOptions()
	snapshot := db.connector.NewSnapshot()
	readOptions.SetSnapshot(snapshot)
	for i, key := range keys {
		value, _ := db.connector.Get(readOptions, key)
		data[i] = value
	}
	db.connector.ReleaseSnapshot(snapshot)
	return data, nil
}

func (db *db) dbRange(start []byte, end []byte) ([][]byte, error) {
	var data [][]byte

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
    if glog.V(8) {
        glog.Infof("Fetched %q from range %q to %q in %q",
            data, start, end, db.Name)
    }
	return data, nil
}

func (db *db) slice(start []byte, limit int) ([][]byte, error) {
	var data [][]byte
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

func (db *db) batch(puts []*BatchPut, deletes []*BatchDelete) error {
	batch := leveldb.NewWriteBatch()
    for _, put := range puts {
        batch.Put(put.Key, put.Value)
    }
    for _, d := range deletes {
        batch.Delete(d.Key)
    }
	wo := leveldb.NewWriteOptions()
    err := db.connector.Write(wo, batch)
	if err != nil {
		return ValueError(err)
	}
	return nil
}
