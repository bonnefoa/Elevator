package elevator

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	"os"
	"bytes"
)

const TestDb = "test_db"

type Tester interface {
        Fatal(args ...interface{})
        Fatalf(format string, args ...interface{})
}

var TestConf = GetTestConf()

func GetTestConf() *Config {
	core := &CoreConfig{
		Daemon:      false,
		Endpoint:    "tcp://127.0.0.1:4141",
		Pidfile:     "test/elevator.pid",
		StorePath:   "test/elevator/store",
		StoragePath: "test/elevator",
		DefaultDb:   "default",
		LogFile:     "test/elevator.log",
		LogLevel:    "INFO",
	}
	storage := NewStorageEngineConfig()
	config := &Config{
		Core:    core,
		Storage: storage,
	}
	l4g.AddFilter("stdout", l4g.INFO, l4g.NewConsoleLogWriter())
	return config
}

func isStringSliceEquals(slc1 []string, slc2 []string) bool {
	if len(slc1) != len(slc2) {
		return false
	}
	for i := range slc1 {
		el1 := slc1[i]
		el2 := slc2[i]
		if el1 != el2 {
			return false
		}
	}
	return true
}

func fillNKeys(db *Db, n int) {
	req := make([]string, n*3)
	for i := 0; i < n*3; i += 3 {
		req[i] = SIGNAL_BATCH_PUT
		req[i+1] = fmt.Sprintf("key_%i", i)
		req[i+2] = fmt.Sprintf("val_%i", i)
	}
	Batch(db, &Request{Args: req})
}

func CleanDbStorage() {
	os.RemoveAll(TestConf.Core.StoragePath)
	os.MkdirAll(TestConf.Core.StoragePath, 0700)
}

func GetTestDb() (*DbStore, *Db, error) {
	CleanDbStorage()
	db_store := NewDbStore(TestConf)
	err := db_store.Add(TestDb)
	if err != nil {
		return nil, nil, err
	}
	request_connect := &Request{Args: []string{TestDb}}
	response_connect, _ := DbConnect(db_store, request_connect)
	if response_connect.Status != SUCCESS_STATUS {
		return nil, nil,
			errors.New(fmt.Sprintf("Error on connection %v",
				response_connect.Status))
	}
	db_uid := response_connect.Data[0]
	db := db_store.Container[db_uid]
	if db == nil {
		return nil, nil,
			errors.New(fmt.Sprintf("No db for uid %v", db_uid))
	}
	if db.Status == DB_STATUS_UNMOUNTED {
		return nil, nil,
			errors.New(fmt.Sprintf("Db is unmounted %s", db_uid))
	}
	return db_store, db, nil
}

func TemplateDbTest(fatalf func(string, ...interface{}), f func(*DbStore, *Db)) {
	db_store, db, err := GetTestDb()
	defer db_store.UnmountAll()
	if err != nil {
		fatalf("Error when creating test db %v", err)
	}
	f(db_store, db)
}

func receiveResponse(t Tester, socket *zmq.Socket) Response {
	var response Response
	parts, _ := socket.RecvMultipart(0)
        if len(parts) == 0 {
                t.Fatal("Received empty response")
        }
	UnpackFrom(&response, bytes.NewBuffer(parts[0]))
	return response
}

func TemplateServerTest(t Tester, f func(*zmq.Socket, string)) {
	CleanDbStorage()
	exitSignal := make(chan bool)
	go ListenAndServe(TestConf, exitSignal)
	defer func() {
		exitSignal<-true
                <-exitSignal
	}()

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	socket.Connect(TestConf.Core.Endpoint)

	sendRequest(Request{Command: DB_CREATE, Args: []string{TestDb}}, socket)
	response := receiveResponse(t, socket)
	if response.Status != SUCCESS_STATUS {
		t.Fatalf("Error on db creation %v", response)
	}
	sendRequest(Request{Command: DB_CONNECT, Args: []string{TestDb}}, socket)
	response = receiveResponse(t, socket)
	if response.Status != SUCCESS_STATUS {
		t.Fatalf("Error on db connection %q", response)
	}

	uid := response.Data[0]
	f(socket, uid)
}
