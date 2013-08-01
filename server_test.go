package elevator

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"testing"
)

func TestPackUnpackRequest(t *testing.T) {
	var buffer bytes.Buffer
	var resultRequest Request

	startRequest := Request{Command: DB_CONNECT, Args: []string{TestDb}}
	PackInto(startRequest, &buffer)

	UnpackFrom(&resultRequest, &buffer)
	if resultRequest.Command != DB_CONNECT {
		t.Fatalf("Expected request to be DB_CONNECT, got %q", resultRequest)
	}
	UnpackFrom(resultRequest, &buffer)
	if resultRequest.Command != DB_CONNECT {
		t.Fatalf("Expected request to be DB_CONNECT, got %q", resultRequest)
	}

	PackInto(&startRequest, &buffer)
	UnpackFrom(&resultRequest, &buffer)
	if resultRequest.Command != DB_CONNECT {
		t.Fatalf("Expected request to be DB_CONNECT, got %q", resultRequest)
	}
	UnpackFrom(resultRequest, &buffer)
	if resultRequest.Command != DB_CONNECT {
		t.Fatalf("Expected request to be DB_CONNECT, got %q", resultRequest)
	}
}

func sendRequest(request Request, socket *zmq.Socket) {
	var buffer bytes.Buffer
	PackInto(request, &buffer)
	socket.SendMultipart([][]byte{buffer.Bytes()}, 0)
}

func TestServer(t *testing.T) {
	f := func(socket *zmq.Socket, uid string) {
		sendRequest(Request{Command: DB_PUT, Args: []string{"key", "val"}, DbUid: uid}, socket)
		response := receiveResponse(t, socket)
		if response.Status != SUCCESS_STATUS {
			t.Fatalf("Error on db put %q", response)
		}

		sendRequest(Request{Command: DB_GET, Args: []string{"key"}, DbUid: uid}, socket)
		response = receiveResponse(t, socket)
		if response.Status != SUCCESS_STATUS {
			t.Fatalf("Error on db get %q", response)
		}
		if response.Data[0] != "val" {
			t.Fatalf("Expected to fetch 'key' value 'val', got %q", response.Data[0])
		}
	}
	TemplateServerTest(t, f)
}

func BenchmarkServerGet(b *testing.B) {
	f := func(socket *zmq.Socket, uid string) {
		args := make([]string, b.N*3)
                b.Logf("b.N is %d", b.N)
		for i := 0; i < b.N*3; i += 3 {
			args[i] = SIGNAL_BATCH_PUT
			args[i+1] = fmt.Sprintf("key_%d", i)
			args[i+2] = fmt.Sprintf("val_%d", i)
		}
		sendRequest(Request{Command: DB_BATCH, Args: args, DbUid: uid}, socket)
		response := receiveResponse(b, socket)
		if response.Status != SUCCESS_STATUS {
			b.Fatalf("Error on db batch %q", response)
		}
                b.Logf("Finished writing")
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                        sendRequest(Request{Command: DB_GET,
                                Args: []string{fmt.Sprintf("key_%d", i)}, DbUid: uid}, socket)
                        response = receiveResponse(b, socket)
                }
                b.Logf("Finished %d queries", b.N)
	}
	TemplateServerTest(b, f)
}
