package store

import (
	"fmt"
	"testing"
	"reflect"
)


func putRequest(key string, val string) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_PUT
    r.Command = &cmd
    r.Put = &PutRequest{[]byte(key), []byte(val), nil}
    return r
}

func getRequest(key string) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_GET
    r.Command = &cmd
    r.Get = &GetRequest{[]byte(key), nil}
    return r
}

func deleteRequest(key string) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_DELETE
    r.Command = &cmd
    r.Delete = &DeleteRequest{[]byte(key), nil}
    return r
}

func mgetRequest(keys ... string) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_MGET
    r.Command = &cmd
    r.Mget = &MgetRequest{ToBytes(keys...), nil}
    return r
}

func rangeRequest(start string, end string) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_RANGE
    r.Command = &cmd
    r.Range = &RangeRequest{[]byte(start), []byte(end), nil}
    return r
}

func sliceRequest(start string, limit int32) DbRequest {
    r := DbRequest{}
    cmd := DbRequest_SLICE
    r.Command = &cmd
    r.Slice = &SliceRequest{[]byte(start), &limit, nil}
    return r
}

func batchRequest(putKeys [][]byte, putsValues[][]byte,
            deleteKeys [][]byte) DbRequest {
    r := DbRequest{}
    batchPuts := make([]*BatchPut, len(putKeys))
    for i, k := range putKeys {
        batchPuts[i] = &BatchPut{k, putsValues[i], nil}
    }
    batchDeletes := make([]*BatchDelete, len(deleteKeys))
    for i, k := range deleteKeys {
        batchDeletes[i] = &BatchDelete{k, nil}
    }
    cmd := DbRequest_BATCH
    r.Command = &cmd
    r.Batch = & BatchRequest { batchPuts, batchDeletes, nil }
    return r
}

var testOperationDatas = []struct {
    r DbRequest
    expectedError   error
    data           [][]byte
}{
    {getRequest("key"), KeyError("key"), nil},

    {putRequest("key", "val"), nil, nil},
    {putRequest("key2", "val2"), nil, nil},
    {putRequest("key3", "val3"), nil, nil},

    {getRequest("key"), nil, ToBytes("val")},
    {mgetRequest("key", "key2"), nil, ToBytes("val", "val2")},
    {rangeRequest("key", "key3"), nil,
        ToBytes("key", "val", "key2", "val2", "key3", "val3")},

	{sliceRequest("key", 2), nil,
		ToBytes("key", "val", "key2", "val2")},

	{deleteRequest("key"), nil, nil},
	{getRequest("key"), KeyError("key"), nil},

	{batchRequest(ToBytes("batch1", "batch2"), ToBytes("val1", "val2"), nil),
        nil, nil},
	{getRequest("batch1"), nil, ToBytes("val1")},

	{batchRequest(ToBytes("batch3"), ToBytes("val3"), ToBytes("batch3")), nil, nil},
	{getRequest("batch3"), KeyError("batch3"), nil},
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
