package store

import (
	"fmt"
	"testing"
	"reflect"
)

var testOperationDatas = []struct {
	op             string
	request        [][]byte
	expectedError   error
	data           [][]byte
}{
	{DbGet, ToBytes("key"), KeyError("key"), nil},

	{DbPut, ToBytes("key", "val"), nil, nil},
	{DbPut, ToBytes("key2", "val2"), nil, nil},
	{DbPut, ToBytes("key3", "val3"), nil, nil},

	{DbGet, ToBytes("key"), nil, ToBytes("val")},
	{DbMget, ToBytes("key", "key2"), nil,
		ToBytes("val", "val2")},
	{DbRange, ToBytes("key", "key3"), nil,
		ToBytes("key", "val", "key2", "val2", "key3", "val3")},

	{DbSlice, ToBytes("key", "2"), nil,
		ToBytes("key", "val", "key2", "val2")},

	{DbDelete, ToBytes("key"), nil, nil},
	{DbGet, ToBytes("key"), KeyError("key"), nil},

	{DbBatch, ToBytes(SignalBatchPut, "batch1", "val1",
		SignalBatchPut, "batch2", "val2"), nil, nil},
	{DbGet, ToBytes("batch1"), nil, ToBytes("val1")},

	{DbBatch, ToBytes(SignalBatchPut, "batch3", "val3",
		SignalBatchDelete, "batch3"), nil, nil},
	{DbGet, ToBytes("batch3"), KeyError("batch3"), nil},
}

func TestOperations(t *testing.T) {
    env := setupEnv(t)
    defer env.destroy()
    for i, tt := range testOperationDatas {
        command := databaseComands[tt.op]
        res, err := command(env.db, tt.request)
        if !reflect.DeepEqual(err, tt.expectedError) {
            t.Fatalf("%d, expected status %v, got %v", i,
            tt.expectedError, err)
        }
        if !reflect.DeepEqual(res, tt.data) {
            t.Fatalf("expected %v, got %v", tt.data, res)
        }
    }
}

func BenchmarkAtomicPut(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        args := ToBytes(fmt.Sprintf("key_%d", i), fmt.Sprintf("val_%d", i))
        put(env.db, args)
    }
}

func BenchmarkBatchPut(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    b.ResetTimer()
    fillNKeys(env.db, b.N)
}

func BenchmarkGet(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    fillNKeys(env.db, b.N)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        args := ToBytes(fmt.Sprintf("key_%d", i))
        get(env.db, args)
    }
}

func BenchmarkBatchDelete(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    fillNKeys(env.db, b.N)
    b.ResetTimer()
    args := make([]string, b.N*2)
    for i := 0; i < b.N*2; i += 2 {
        args[i] = SignalBatchDelete
        args[i+1] = fmt.Sprintf("key_%d", i)
    }
    batch(env.db, ToBytes(args...))
}

func BenchmarkDelete(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    fillNKeys(env.db, b.N)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        args := ToBytes(fmt.Sprintf("key_%d", i))
        dbDelete(env.db, args)
    }
}

func templateMGet(b *testing.B, numKeys int, fun func(*db, [][]byte) ([][]byte, error)) {
    env := setupEnv(b)
    defer env.destroy()
    fillNKeys(env.db, b.N)
    get := make([][]byte, numKeys)
    for i := 0; i < numKeys; i++ {
        get[i] = []byte(fmt.Sprintf("key_%d", i))
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        fun(env.db, get)
    }
}

func Benchmark10MGet(b *testing.B) {
	templateMGet(b, 10, mGet)
}
