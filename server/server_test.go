package server

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"testing"
	store "github.com/oleiade/Elevator/store"
    "reflect"
)

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type Env struct {
	*zmq.Context
	client     *zmq.Socket
	endpoint   string
	Tester
}

func (env *Env) setupEnv() {
    var err error
    env.Context, err = zmq.NewContext()
    if err != nil {
        env.Fatalf("Error on context creation", err)
    }
	socket, err := env.NewSocket(zmq.REQ)
    if err != nil {
        env.Fatalf("Error on socket creation", err)
    }
    err = socket.Connect(env.endpoint)
    if err != nil {
        env.Fatalf("Error on socket connect", err)
    }

	c := getTestConf()
	defer cleanConfTemp(c)
	exitSignal := make(chan bool)
	go ListenAndServe(c, exitSignal)
	defer func() {
		exitSignal <- true
		<-exitSignal
	}()


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


func TestServer(t *testing.T) {
	f := func(socket *zmq.Socket, uid string) {
		req := store.Request{Command: store.DB_PUT,
            Args: store.ToBytes("key", "val"), DbUid: uid}
		req.SendRequest(socket)
		response := ReceiveResponse(socket)
		if response.Status != SUCCESS {
			t.Fatalf("Error on db put %q", response)
		}

		req = store.Request{Command: store.DB_GET,
            Args: store.ToBytes("key"), DbUid: uid}
		req.SendRequest(socket)
		response = ReceiveResponse(socket)
		if response.Status != SUCCESS {
			t.Fatalf("Error on db get %q", response)
		}
        if !reflect.DeepEqual(response.Data, []byte("val")) {
			t.Fatalf("Expected to fetch 'key' value 'val', got %q",
                response.Data[0])
		}
	}
	TemplateServerTest(t, f)
}

func BenchmarkServerGet(b *testing.B) {
	f := func(socket *zmq.Socket, uid string) {
		args := make([]string, b.N*3)
		b.Logf("b.N is %d", b.N)
		for i := 0; i < b.N*3; i += 3 {
			args[i] = store.SIGNAL_BATCH_PUT
			args[i+1] = fmt.Sprintf("key_%d", i)
			args[i+2] = fmt.Sprintf("val_%d", i)
		}
		req := store.Request{Command: store.DB_BATCH,
            Args: store.ToBytes(args...), DbUid: uid}
		req.SendRequest(socket)
		response := ReceiveResponse(socket)
		if response.Status != SUCCESS {
			b.Fatalf("Error on db batch %q", response)
		}
		b.Logf("Finished writing")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req = store.Request{Command: store.DB_GET,
				Args: store.ToBytes(fmt.Sprintf("key_%d", i)), DbUid: uid}
			req.SendRequest(socket)
			response = ReceiveResponse(socket)
		}
		b.StopTimer()
		b.Logf("Finished %d queries", b.N)
	}
	TemplateServerTest(b, f)
}
