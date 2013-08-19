package store

import (
	"fmt"
	"testing"
	"reflect"
)


var (
    key1 = []byte("key1")
    key2 = []byte("key2")
    key3 = []byte("key3")
    val1 = []byte("val1")
    val2 = []byte("val2")
    val3 = []byte("val3")
    batch1 = []byte("batch1")
    batch2 = []byte("batch2")
    batch3 = []byte("batch3")
)

var testOperationDatas = []struct {
    r DbRequest
    expectedError   error
    data           [][]byte
}{
    {NewGetRequest(TestDb, key1), KeyError(key1), nil},

    {NewPutRequest(TestDb, key1, val1), nil, nil},
    {NewPutRequest(TestDb,key2,val2), nil, nil},
    {NewPutRequest(TestDb, key3, val3), nil, nil},

    {NewGetRequest(TestDb, key1), nil, [][]byte{val1}},
    {NewMgetRequest(TestDb, key1,key2), nil, [][]byte{val1,val2}},
    {NewRangeRequest(TestDb, key1, key3), nil,
        [][]byte{key1, val1, key2, val2, key3, val3}},

	{NewSliceRequest(TestDb, key1, 2), nil,
		[][]byte{key1, val1,key2,val2}},

	{NewDeleteRequest(TestDb, key1), nil, nil},
	{NewGetRequest(TestDb, key1), KeyError(key1), nil},

	{NewBatchRequest(TestDb, [][]byte{batch1, batch2}, [][]byte{val1,val2}, nil),
        nil, nil},
	{NewGetRequest(TestDb, batch1), nil, [][]byte{val1}},

	{NewBatchRequest(TestDb, [][]byte{batch3}, [][]byte{val3},
        [][]byte{batch3}), nil, nil},
	{NewGetRequest(TestDb, batch3), KeyError(batch3), nil},
}

func TestOperations(t *testing.T) {
    env := setupEnv(t)
    defer env.destroy()
    for i, tt := range testOperationDatas {
        dbName := TestDb
        tt.r.DbName = &dbName
        res, err := env.db.processRequest(&tt.r)
        if !reflect.DeepEqual(err, tt.expectedError) {
            t.Fatalf("%d: expected status %v, got %v", i,
            tt.expectedError, err)
        }
        if !reflect.DeepEqual(res, tt.data) {
            t.Fatalf("%d: expected %q, got %q", i, tt.data, res)
        }
    }
}

func BenchmarkAtomicPut(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := []byte(fmt.Sprintf("key_%d", i))
        value := []byte(fmt.Sprintf("val_%d", i))
        env.db.put(key, value)
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
        key := []byte(fmt.Sprintf("key_%d", i))
        env.db.get(key)
    }
}

func BenchmarkBatchDelete(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    fillNKeys(env.db, b.N)
    b.ResetTimer()
    deletes := make([]*BatchDelete, b.N)
    for i := 0; i < b.N; i++ {
        deletes[i] = &BatchDelete{[]byte(fmt.Sprintf("key_%d", i)), nil}
    }
    env.db.batch(nil, deletes)
}

func BenchmarkDelete(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()

    fillNKeys(env.db, b.N)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        key := []byte(fmt.Sprintf("key_%d", i))
        env.db.dbDelete(key)
    }
}

func Benchmark10Mget(b *testing.B) {
    env := setupEnv(b)
    defer env.destroy()
    fillNKeys(env.db, b.N)
    keys := make([][]byte, 10)
    for i := 0; i < 10; i++ {
        keys[i] = []byte(fmt.Sprintf("key_%d", i))
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        env.db.mget(keys)
    }
}
