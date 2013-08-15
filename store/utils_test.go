package store

import (
	"testing"
	"bytes"
)

func BenchmarkPackUnpack(b *testing.B) {
	args := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		args[i] = []byte(string(i))
	}
	request := Request{Command: DbGet, Args: args, DbUID: "Test"}

	var buffer bytes.Buffer
	PackInto(request, &buffer)

	var resultRequest Request
	UnpackFrom(resultRequest, &buffer)
}

func TestPackUnpackRequest(t *testing.T) {
	var buffer bytes.Buffer
	var resultRequest Request

	startRequest := Request{Command: DbConnect, Args: ToBytes(TestDb)}
	PackInto(startRequest, &buffer)

	UnpackFrom(&resultRequest, &buffer)
	if resultRequest.Command != DbConnect {
		t.Fatalf("Expected request to be DbConnect, got %q", resultRequest)
	}
	UnpackFrom(resultRequest, &buffer)
	if resultRequest.Command != DbConnect {
		t.Fatalf("Expected request to be DbConnect, got %q", resultRequest)
	}

	PackInto(&startRequest, &buffer)
	UnpackFrom(&resultRequest, &buffer)
	if resultRequest.Command != DbConnect {
		t.Fatalf("Expected request to be DbConnect, got %q", resultRequest)
	}
	UnpackFrom(resultRequest, &buffer)
	if resultRequest.Command != DbConnect {
		t.Fatalf("Expected request to be DbConnect, got %q", resultRequest)
	}
}

