package elevator

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	"io/ioutil"
	"os"
)

const TestDb = "test_db"

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func getTestConf() *Config {
	logConfig := &LogConfiguration{
		LogFile:  "test/elevator.log",
		LogLevel: "INFO",
	}
	storePath, _ := ioutil.TempFile("/tmp", "elevator_store")
	storagePath, _ := ioutil.TempDir("/tmp", "elevator_path")

	core := &CoreConfig{
		Daemon:      false,
		Endpoint:    "tcp://127.0.0.1:4141",
		Pidfile:     "test/elevator.pid",
		StorePath:   storePath.Name(),
		StoragePath: storagePath,
		DefaultDb:   "default",
	}
	storage := NewStorageEngineConfig()
	options := storage.ToLeveldbOptions()
	config := &Config{
		core,
		storage,
		logConfig,
		options,
	}
	l4g.AddFilter("stdout", l4g.INFO, l4g.NewConsoleLogWriter())
	return config
}

func cleanConfTemp(c *Config) {
	os.RemoveAll(c.StoragePath)
	os.RemoveAll(c.StorePath)
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

func TemplateDbTest(t Tester, f func(*DbStore, *Db)) {
	c := getTestConf()
	defer cleanConfTemp(c)
	db_store := NewDbStore(c)
	err := db_store.Add(TestDb)
	if err != nil {
		t.Fatalf("Error when creating test db %v", err)
	}
	request_connect := &Request{Args: []string{TestDb}}
	response_connect, _ := DbConnect(db_store, request_connect)
	if response_connect.Status != SUCCESS_STATUS {
		t.Fatalf("Error on connection %v",
			response_connect.Status)
	}
	db_uid := response_connect.Data[0]
	db := db_store.Container[db_uid]
	if db == nil {
		t.Fatalf("No db for uid %v", db_uid)
	}
	if db.Status == DB_STATUS_UNMOUNTED {
		t.Fatalf("Db is unmounted %s", db_uid)
	}

	defer db_store.UnmountAll()
	if err != nil {
		t.Fatalf("Error when creating test db %v", err)
	}
	f(db_store, db)
}

func TemplateServerTest(t Tester, f func(*zmq.Socket, string)) {
	c := getTestConf()
	defer cleanConfTemp(c)
	exitSignal := make(chan bool)
	go ListenAndServe(c, exitSignal)
	defer func() {
		exitSignal <- true
		<-exitSignal
	}()

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	socket.Connect(c.Endpoint)

	req := Request{Command: DB_CREATE, Args: []string{TestDb}}
	req.SendRequest(socket)
	response := ReceiveResponse(socket)
	if response.Status != SUCCESS_STATUS {
		t.Fatalf("Error on db creation %v (test conf was %q)",
			response, c)
	}
	req = Request{Command: DB_CONNECT, Args: []string{TestDb}}
	req.SendRequest(socket)
	response = ReceiveResponse(socket)
	if response.Status != SUCCESS_STATUS {
		t.Fatalf("Error on db connection %q", response)
	}

	uid := response.Data[0]
	f(socket, uid)
}
