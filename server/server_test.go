package server

import (
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	store "github.com/oleiade/Elevator/store"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

const (
	TestDb       = "test_db"
	TestEndpoint = "tcp://127.0.0.1:4141"
)

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type Env struct {
	*zmq.Context
	*zmq.Socket
	Tester
	uid string
	*Config
	exitChannel chan bool
}

func cleanConfTemp(c *store.StoreConfig) {
	os.RemoveAll(c.StorePath)
	os.RemoveAll(c.StoragePath)
}

func getTestConf() *Config {
	tempDir, _ := ioutil.TempDir("/tmp", "elevator")
	config := NewConfig()
	logConfig := &LogConfiguration{
		LogFile:  path.Join(tempDir, "elevator.log"),
		LogLevel: "INFO",
	}
	ConfigureLogger(logConfig)

	config.StoreConfig.CoreConfig.StorePath = path.Join(tempDir, "store")
	config.StoreConfig.CoreConfig.StoragePath = tempDir

	config.ServerConfig.Endpoint = TestEndpoint
	config.ServerConfig.Pidfile = path.Join(tempDir, "elevator.pid")

	config.LogConfiguration = logConfig
	return config
}

func (env *Env) setupEnv() {
	var err error
	env.Context, err = zmq.NewContext()
	if err != nil {
		env.Fatalf("Error on context creation", err)
	}
	env.Socket, err = env.NewSocket(zmq.REQ)
	if err != nil {
		env.Fatalf("Error on socket creation", err)
	}
	err = env.Socket.Connect(TestEndpoint)
	if err != nil {
		env.Fatalf("Error on socket connect", err)
	}
	env.Config = getTestConf()
	env.exitChannel = make(chan bool)
	go ListenAndServe(env.Config, env.exitChannel)
	req := store.Request{Command: store.DB_CREATE, Args: store.ToBytes(TestDb)}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != SUCCESS {
		env.Fatalf("Error on db creation %v (test conf was %q)",
			response, env.Config)
	}
	req = store.Request{Command: store.DB_CONNECT, Args: store.ToBytes(TestDb)}
	req.SendRequest(env.Socket)
	response = ReceiveResponse(env.Socket)
	if response.Status != SUCCESS {
		env.Fatalf("Error on db connection %q", response)
	}
	env.uid = string(response.Data[0])
}

func (env *Env) destroy() {
	env.exitChannel<- true
	<-env.exitChannel
	env.CleanConfiguration()
}

func TestServer(t *testing.T) {
	env := Env{Tester: t}
	env.setupEnv()
	defer env.destroy()
	req := store.Request{Command: store.DB_PUT, Args: store.ToBytes("key", "val"), DbUid: env.uid}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != SUCCESS {
		t.Fatalf("Error on db put %q", response)
	}
	req = store.Request{Command: store.DB_GET,
	Args: store.ToBytes("key"), DbUid: env.uid}
	req.SendRequest(env.Socket)
	response = ReceiveResponse(env.Socket)
	if response.Status != SUCCESS {
		t.Fatalf("Error on db get %q", response)
	}
	expectedValue := store.ToBytes("val")
	if !reflect.DeepEqual(response.Data, expectedValue) {
		t.Fatalf("Expected to fetch 'key' value %q, got %q", expectedValue, response.Data[0])
	}
}

func BenchmarkServerGet(b *testing.B) {
	env := Env{Tester: b}
	env.setupEnv()
	defer env.destroy()

	args := make([]string, b.N*3)
	b.Logf("b.N is %d", b.N)
	for i := 0; i < b.N*3; i += 3 {
		args[i] = store.SIGNAL_BATCH_PUT
		args[i+1] = fmt.Sprintf("key_%d", i)
		args[i+2] = fmt.Sprintf("val_%d", i)
	}
	req := store.Request{Command: store.DB_BATCH, Args: store.ToBytes(args...), DbUid: env.uid}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != SUCCESS {
		b.Fatalf("Error on db batch %q", response)
	}
	b.Logf("Finished writing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req = store.Request{Command: store.DB_GET,
			Args: store.ToBytes(fmt.Sprintf("key_%d", i)), DbUid: env.uid}
		req.SendRequest(env.Socket)
		response = ReceiveResponse(env.Socket)
	}
	b.StopTimer()
	b.Logf("Finished %d queries", b.N)
}
