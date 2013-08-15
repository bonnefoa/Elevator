package store

import (
	"testing"
)

func TestError(t *testing.T) {
    key := []byte("test")
    err := KeyError(key)
    if err.Error() != "Key \"test\" does not exists" {
        t.Fatalf("Got errors")
    }
}
