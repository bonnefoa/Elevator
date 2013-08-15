package store

import (
	uuid "code.google.com/p/go-uuid/uuid"
	"fmt"
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
	requestChan  chan *Request   `json:"-"`
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
func processRequest(db *db, command string, args [][]byte) ([][]byte, error) {
	f, ok := databaseComands[command]
	if !ok {
		err := RequestError( fmt.Errorf("Unknown command %s", command) )
		return nil, err
	}
	response, err := f(db, args)
	return response, err
}

// StartRoutine listens on the Db channel awaiting
// for incoming requests to execute. Willingly
// blocking on each Exec call received through the
// channel in order to protect requests.
func (db *db) StartRoutine() {
	for request := range db.requestChan {
		res, err := processRequest(db, request.Command, request.Args)
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
	if db.status == statusUnmounted {
		db.connector, err = leveldb.Open(db.Path, options)
		if err != nil {
			return err
		}
		db.status = statusMounted
		db.requestChan = make(chan *Request, 100)
		db.responseChan = make(chan *dbResult, 100)
		go db.StartRoutine()
	} else {
		err := fmt.Errorf("Database %s already mounted", db.Name)
		glog.Error(err)
		return err
	}
	if glog.V(2) {
        glog.Info(fmt.Sprintf("Database %s mounted", db.Name))
	}
	return nil
}

// Unmount sets the database status to statusUnmounted
// and deletes the according leveldb connector
func (db *db) Unmount() (err error) {
	if db.status == statusMounted {
		db.connector.Close()
		close(db.requestChan)
		close(db.responseChan)
		db.status = statusUnmounted
	} else {
        err := fmt.Errorf("Db: Database %s already unmounted", db.Name)
		glog.Error(err)
		return err
	}
	if glog.V(2) {
        glog.Infoln("Db: Database %s unmounted", db.Name)
    }
	return nil
}
