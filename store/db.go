package store

import (
	uuid "code.google.com/p/go-uuid/uuid"
	"fmt"
	l4g "github.com/alecthomas/log4go"
	leveldb "github.com/jmhodges/levigo"
)

type DbResult struct {
	Data [][]byte
	Err error
}

type Db struct {
	Name         string          `json:"name"`
	Uid          string          `json:"uid"`
	Path         string          `json:"path"`
	status       DbMountedStatus `json:"-"`
	connector    *leveldb.DB     `json:"-"`
	requestChan  chan *Request   `json:"-"`
	responseChan chan *DbResult  `json:"-"`
}

func NewDb(db_name string, path string) *Db {
	return &Db{
		Name:         db_name,
		Path:         path,
		Uid:          uuid.New(),
		status:       DB_STATUS_UNMOUNTED,
		requestChan:  make(chan *Request),
		responseChan: make(chan *DbResult),
	}
}

// processRequest executes the received request command, and returns
// the resulting response.
func processRequest(db *Db, command string, args [][]byte) ([][]byte, error) {
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
func (db *Db) StartRoutine() {
	for request := range db.requestChan {
		res, err := processRequest(db, request.Command, request.Args)
		if err != nil {
			l4g.Error(err)
		}
		result := &DbResult{
			Data:res,
			Err:err,
		}
		db.responseChan <- result
	}
}

// Mount sets the database status to DB_STATUS_MOUNTED
// and instantiates the according leveldb connector
func (db *Db) Mount(options *leveldb.Options) (err error) {
	if db.status == DB_STATUS_UNMOUNTED {
		db.connector, err = leveldb.Open(db.Path, options)
		if err != nil {
			return err
		}
		db.status = DB_STATUS_MOUNTED
		db.requestChan = make(chan *Request, 100)
		db.responseChan = make(chan *DbResult, 100)
		go db.StartRoutine()
	} else {
		error := fmt.Errorf("Database %s already mounted", db.Name)
		l4g.Error(error)
		return error
	}
	l4g.Debug(func() string {
		return fmt.Sprintf("Database %s mounted", db.Name)
	})
	return nil
}

// Unmount sets the database status to DB_STATUS_UNMOUNTED
// and deletes the according leveldb connector
func (db *Db) Unmount() (err error) {
	if db.status == DB_STATUS_MOUNTED {
		db.connector.Close()
		close(db.requestChan)
		close(db.responseChan)
		db.status = DB_STATUS_UNMOUNTED
	} else {
		error := fmt.Errorf("Database %s already unmounted", db.Name)
		l4g.Error(error)
		return error
	}
	l4g.Debug(func() string {
		return fmt.Sprintf("Database %s unmounted", db.Name)
	})
	return nil
}
