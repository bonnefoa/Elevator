package store

import (
    "testing"
    "bytes"
)

func TestPackUnpackRequest(t *testing.T) {
	var buffer bytes.Buffer
	var resultRequest Request

	startRequest := Request{Command: DB_CONNECT, Args: ToBytes(TestDb)}
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

