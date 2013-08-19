package server

import (
	"code.google.com/p/goprotobuf/proto"
	"fmt"
	zmq "github.com/bonnefoa/go-zeromq"
	store "github.com/oleiade/Elevator/store"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var (
	key1 = []byte("key1")
	key2 = []byte("key2")
	val1 = []byte("val1")
)

const (
	TestEndpoint = "tcp://127.0.0.1:4141"
	TestDb       = "test_db"
)

type Tester interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type Env struct {
	*zmq.Context
	*zmq.Socket
	Tester
	*Config
	exitChannel chan bool
}

func cleanConfTemp(c *store.StoreConfig) {
	os.RemoveAll(c.StorePath)
	os.RemoveAll(c.StoragePath)
}

func getTestConf() *Config {
	tempDir, _ := ioutil.TempDir("/tmp", "elevator")
	config := newConfig()

	config.StoreConfig.CoreConfig.StorePath = path.Join(tempDir, "store")
	config.StoreConfig.CoreConfig.StoragePath = tempDir

	config.serverConfig.Endpoint = TestEndpoint
	config.serverConfig.Pidfile = path.Join(tempDir, "elevator.pid")

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
	env.Config = getTestConf()
	env.exitChannel = make(chan bool)
	go ListenAndServe(env.Config, env.exitChannel)

	// Create base
	createReq := store.NewStoreRequest(TestDb, store.StoreRequest_CREATE)
	SendRequest(&createReq, env.Socket)
	response, err := ReceiveResponse(env.Socket)
	if err != nil {
		t.Fatalf("Error on send create db request %q", err)
	}
	if *response.Status != Response_SUCCESS {
		env.Fatalf("Error on db creation %v (test conf was %q)",
			response, env.Config)
	}

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
	// Test put
	dbReq := store.NewPutRequest(TestDb, key1, val1)
	err := SendDbRequest(&dbReq, env.Socket)
	if err != nil {
		t.Fatalf("Error on send get request %q", err)
	}
	response, err := ReceiveResponse(env.Socket)
	if err != nil {
		t.Fatalf("Error on receive get request %q", err)
	}
	if *response.Status != Response_SUCCESS {
		t.Fatalf("Error on db put %q", response)
	}

	// Test get
	dbReq = store.NewGetRequest(TestDb, key1)
	SendDbRequest(&dbReq, env.Socket)
	response, err = ReceiveResponse(env.Socket)
	if err != nil {
		t.Fatal("Error on get response receive", err)
	}
	if *response.Status != Response_SUCCESS {
		t.Fatalf("Error on db get %q", response)
	}
	expectedValue := [][]byte{val1}
	if !reflect.DeepEqual(response.Data, expectedValue) {
		t.Fatalf("Expected to fetch 'key' value %q, got %q",
			expectedValue, response.Data[0])
	}

	// Test get on unknown key
	dbReq = store.NewGetRequest(TestDb, key2)
	SendDbRequest(&dbReq, env.Socket)
	response, err = ReceiveResponse(env.Socket)
	if err != nil {
		t.Fatal("Error on get response receive", err)
	}
	if *response.Status != Response_KEY_ERROR {
		t.Fatalf("Expected key error, got %q", response.Status)
	}
}

func TestBigPut(t *testing.T) {
	env := setupEnv(t)
	defer env.destroy()

	putKeys, putValues := getMPut(30000)
	req := store.NewBatchRequest(TestDb, putKeys, putValues, nil)

	SendDbRequest(&req, env.Socket)
	response, err := ReceiveResponse(env.Socket)
	if err != nil {
		t.Fatal("Error on put request receive", err)
	}
	if *response.Status != Response_SUCCESS {
		t.Fatalf("Error on db put %q", response)
	}
}

func getMPut(n int) ([][]byte, [][]byte) {
	putKeys := make([][]byte, n)
	putValues := make([][]byte, n)
	for i := 0; i < n; i++ {
		putKeys[i] = []byte(fmt.Sprintf("key_%d", i))
		putValues[i] = []byte(fmt.Sprintf("val_%d", i))
	}
	return putKeys, putValues
}

func BenchmarkServerGet(b *testing.B) {
	env := setupEnv(b)
	defer env.destroy()

	putKeys, putValues := getMPut(b.N)
	req := store.NewBatchRequest(TestDb, putKeys, putValues, nil)
	SendDbRequest(&req, env.Socket)
	response, _ := ReceiveResponse(env.Socket)
	if *response.Status != Response_SUCCESS {
		b.Fatalf("Error on db batch %q", response)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("key_%d", i))
		req = store.NewGetRequest(TestDb, key)
		SendDbRequest(&req, env.Socket)
		response, _ = ReceiveResponse(env.Socket)
	}
	b.StopTimer()
}

func BenchmarkServerList(b *testing.B) {
	env := setupEnv(b)
	defer env.destroy()

	r := store.NewStoreRequest("default", store.StoreRequest_LIST)
	data, _ := proto.Marshal(&r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Socket.SendMultipart([][]byte{data}, 0)
		env.Socket.Recv(0)
	}
	b.StopTimer()
}
