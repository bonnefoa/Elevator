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
	{DB_GET, ToBytes("key"), KeyError("key"), nil},

	{DB_PUT, ToBytes("key", "val"), nil, nil},
	{DB_PUT, ToBytes("key2", "val2"), nil, nil},
	{DB_PUT, ToBytes("key3", "val3"), nil, nil},

	{DB_GET, ToBytes("key"), nil, ToBytes("val")},
	{DB_MGET, ToBytes("key", "key2"), nil,
		ToBytes("val", "val2")},
	{DB_RANGE, ToBytes("key", "key3"), nil,
		ToBytes("key", "val", "key2", "val2", "key3", "val3")},

	{DB_SLICE, ToBytes("key", "2"), nil,
		ToBytes("key", "val", "key2", "val2")},

	{DB_DELETE, ToBytes("key"), nil, nil},
	{DB_GET, ToBytes("key"), KeyError("key"), nil},

	{DB_BATCH, ToBytes(SIGNAL_BATCH_PUT, "batch1", "val1",
		SIGNAL_BATCH_PUT, "batch2", "val2"), nil, nil},
	{DB_GET, ToBytes("batch1"), nil, ToBytes("val1")},

	{DB_BATCH, ToBytes(SIGNAL_BATCH_PUT, "batch3", "val3",
		SIGNAL_BATCH_DELETE, "batch3"), nil, nil},
	{DB_GET, ToBytes("batch3"), KeyError("batch3"), nil},
}

func TestOperations(t *testing.T) {
	f := func(db_store *DbStore, db *Db) {
		for i, tt := range testOperationDatas {
			command := databaseComands[tt.op]
			res, err := command(db, tt.request)
			if !reflect.DeepEqual(err, tt.expectedError) {
				t.Fatalf("%d, expected status %v, got %v", i,
				tt.expectedError, err)
			}
			if !reflect.DeepEqual(res, tt.data) {
				t.Fatalf("expected %v, got %v", tt.data, res)
			}
		}
	}
	TemplateDbTest(t, f)
}

func BenchmarkAtomicPut(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			args := ToBytes(fmt.Sprintf("key_%i", i), fmt.Sprintf("val_%i", i))
			Put(db, args)
		}
	}
	TemplateDbTest(b, f)
}

func BenchmarkBatchPut(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		b.ResetTimer()
		fillNKeys(db, b.N)
	}
	TemplateDbTest(b, f)
}

func BenchmarkGet(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			args := ToBytes(fmt.Sprintf("key_%i", i))
			Get(db, args)
		}
	}
	TemplateDbTest(b, f)
}

func BenchmarkBatchDelete(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		b.ResetTimer()
		args := make([]string, b.N*2)
		for i := 0; i < b.N*2; i += 2 {
			args[i] = SIGNAL_BATCH_DELETE
			args[i+1] = fmt.Sprintf("key_%i", i)
		}
		Batch(db, ToBytes(args...))
	}
	TemplateDbTest(b, f)
}

func BenchmarkDelete(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			args := ToBytes(fmt.Sprintf("key_%i", i))
			Delete(db, args)
		}
	}
	TemplateDbTest(b, f)
}

func templateMGet(b *testing.B, numKeys int, fun func(*Db, [][]byte) ([][]byte, error)) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		get := make([][]byte, numKeys)
		for i := 0; i < numKeys; i++ {
			get[i] = []byte(fmt.Sprintf("key_%i", i))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fun(db, get)
		}
	}
	TemplateDbTest(b, f)
}

func Benchmark10MGet(b *testing.B) {
	templateMGet(b, 10, MGet)
}
