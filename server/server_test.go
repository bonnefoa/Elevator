package server

import (
	"fmt"
	zmq "github.com/bonnefoa/go-zeromq"
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
	*config
	exitChannel chan bool
}

func cleanConfTemp(c *store.StoreConfig) {
	os.RemoveAll(c.StorePath)
	os.RemoveAll(c.StoragePath)
}

func getTestConf() *config {
	tempDir, _ := ioutil.TempDir("/tmp", "elevator")
	config := newConfig()
	logConfig := &logConfiguration{
		LogFile:  path.Join(tempDir, "elevator.log"),
		LogLevel: "DEBUG",
	}
	configureLogger(logConfig)

	config.StoreConfig.CoreConfig.StorePath = path.Join(tempDir, "store")
	config.StoreConfig.CoreConfig.StoragePath = tempDir

	config.serverConfig.Endpoint = TestEndpoint
	config.serverConfig.Pidfile = path.Join(tempDir, "elevator.pid")

	config.logConfiguration = logConfig
	return config
}

func setupEnv(t Tester) *Env {
	env := &Env{Tester: t}
	var err error
	env.Context, err = zmq.NewContext()
	if err != nil {
		env.Fatal("Error on context creation", err)
	}
	env.Socket, err = env.NewSocket(zmq.Req)
	if err != nil {
		env.Fatal("Error on socket creation", err)
	}
	err = env.Socket.Connect(TestEndpoint)
	if err != nil {
		env.Fatal("Error on socket connect", err)
	}
	env.config = getTestConf()
	env.exitChannel = make(chan bool)
	go ListenAndServe(env.config, env.exitChannel)
	req := store.Request{Command: store.DbCreate, Args: store.ToBytes(TestDb)}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != Success {
		env.Fatalf("Error on db creation %v (test conf was %q)",
			response, env.config)
	}
	req = store.Request{Command: store.DbConnect, Args: store.ToBytes(TestDb)}
	req.SendRequest(env.Socket)
	response = ReceiveResponse(env.Socket)
	if response.Status != Success {
		env.Fatalf("Error on db connection %q", response)
	}
	env.uid = string(response.Data[0])
	return env
}

func (env *Env) destroy() {
	env.exitChannel <- true
	<-env.exitChannel
	env.CleanConfiguration()
}

func TestServerPutGet(t *testing.T) {
	env := setupEnv(t)
	defer env.destroy()
	req := store.Request{Command: store.DbPut, Args: store.ToBytes("key", "val"), DbUID: env.uid}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != Success {
		t.Fatalf("Error on db put %q", response)
	}
	req = store.Request{Command: store.DbGet,
		Args: store.ToBytes("key"), DbUID: env.uid}
	req.SendRequest(env.Socket)
	response = ReceiveResponse(env.Socket)
	if response.Status != Success {
		t.Fatalf("Error on db get %q", response)
	}
	expectedValue := store.ToBytes("val")
	if !reflect.DeepEqual(response.Data, expectedValue) {
		t.Fatalf("Expected to fetch 'key' value %q, got %q", expectedValue, response.Data[0])
	}

	req = store.Request{Command: store.DbGet, Args: store.ToBytes("key_2"), DbUID: env.uid}
	req.SendRequest(env.Socket)
	response = ReceiveResponse(env.Socket)
	if response.Status != KeyError {
		t.Fatalf("Expected key error, got %q", response.Status)
	}
}

func getMPut(n int) [][]byte {
	args := make([]string, n*3)
	for i := 0; i < n*3; i += 3 {
		args[i] = store.SignalBatchPut
		args[i+1] = fmt.Sprintf("key_%d", i)
		args[i+2] = fmt.Sprintf("val_%d", i)
	}
	return store.ToBytes(args...)
}

func BenchmarkServerGet(b *testing.B) {
	env := setupEnv(b)
	defer env.destroy()

	args := getMPut(b.N)
	req := store.Request{Command: store.DbBatch, Args: args, DbUID: env.uid}
	req.SendRequest(env.Socket)
	response := ReceiveResponse(env.Socket)
	if response.Status != Success {
		b.Fatalf("Error on db batch %q", response)
	}
	b.ResetTimer()
	for i := 0; i < b.N*3; i+=3 {
		request := store.Request{Command: store.DbGet,
			Args: store.ToBytes(fmt.Sprintf("key_%d", i)), DbUID: env.uid}
		request.SendRequest(env.Socket)
		response = ReceiveResponse(env.Socket)
	}
	b.StopTimer()
}
