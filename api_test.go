package elevator

import (
	"bytes"
	"fmt"
	leveldb "github.com/jmhodges/levigo"
	"testing"
)

var test_operations_data = []struct {
	op              string
	request         []string
	expected_status int
	expected_data   []string
}{
	{DB_GET, []string{"key"}, FAILURE_STATUS, []string{}},

	{DB_PUT, []string{"key", "val"}, SUCCESS_STATUS, []string{}},
	{DB_PUT, []string{"key2", "val2"}, SUCCESS_STATUS, []string{}},
	{DB_PUT, []string{"key3", "val3"}, SUCCESS_STATUS, []string{}},

	{DB_GET, []string{"key"}, SUCCESS_STATUS, []string{"val"}},
	{DB_MGET, []string{"key", "key2"}, SUCCESS_STATUS,
		[]string{"val", "val2"}},
	{DB_RANGE, []string{"key", "key3"}, SUCCESS_STATUS,
		[]string{"key", "val", "key2", "val2", "key3", "val3"}},

	{DB_SLICE, []string{"key", "2"}, SUCCESS_STATUS,
		[]string{"key", "val", "key2", "val2"}},

	{DB_DELETE, []string{"key"}, SUCCESS_STATUS, []string{}},
	{DB_GET, []string{"key"}, FAILURE_STATUS, []string{}},

	{DB_BATCH, []string{SIGNAL_BATCH_PUT, "batch1", "val1",
		SIGNAL_BATCH_PUT, "batch2", "val2"}, SUCCESS_STATUS, []string{}},
	{DB_GET, []string{"batch1"}, SUCCESS_STATUS, []string{"val1"}},

	{DB_BATCH, []string{SIGNAL_BATCH_PUT, "batch3", "val3",
		SIGNAL_BATCH_DELETE, "batch3"}, SUCCESS_STATUS, []string{}},
	{DB_GET, []string{"batch3"}, FAILURE_STATUS, []string{}},
}

func TestOperations(t *testing.T) {
	f := func(db_store *DbStore, db *Db) {
		for i, tt := range test_operations_data {
			command := database_commands[tt.op]
			resp, _ := command(db, &Request{Args: tt.request})
			if resp.Status != tt.expected_status {
				t.Fatalf("%d, expected status %v, got %v", i,
					tt.expected_status, resp.Status)
			}
			if len(resp.Data) != len(tt.expected_data) {
				t.Fatalf("%d, expected %q, got %q", i, tt.expected_data,
					resp.Data)
			}
			for j, v := range resp.Data {
				if v != tt.expected_data[j] {
					t.Fatalf("%d, expected %v, got %v", i, tt.expected_data,
						resp.Data)
				}
			}
		}
	}
	TemplateDbTest(t.Fatalf, f)
}

func BenchmarkAtomicPut(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Put(db, &Request{Args: []string{fmt.Sprintf("key_%i", i),
				fmt.Sprintf("val_%i", i)}})
		}
	}
	TemplateDbTest(b.Fatalf, f)
}

func BenchmarkBatchPut(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		b.ResetTimer()
		fillNKeys(db, b.N)
	}
	TemplateDbTest(b.Fatalf, f)
}

func BenchmarkGet(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Get(db, &Request{Args: []string{fmt.Sprintf("key_%i", i)}})
		}
	}
	TemplateDbTest(b.Fatalf, f)
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
		req := &Request{Args: args}
		Batch(db, req)
	}
	TemplateDbTest(b.Fatalf, f)
}

func BenchmarkDelete(b *testing.B) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := &Request{Args: []string{fmt.Sprintf("key_%i", i)}}
			Delete(db, req)
		}
	}
	TemplateDbTest(b.Fatalf, f)
}

func OldMGet(db *Db, request *Request) (*Response, error) {
	var response *Response
	var data []string = make([]string, len(request.Args))
	read_options := leveldb.NewReadOptions()
	snapshot := db.Connector.NewSnapshot()
	read_options.SetSnapshot(snapshot)
	if len(request.Args) > 0 {
		start := request.Args[0]
		end := request.Args[len(request.Args)-1]
		keys_index := make(map[string]int)
		for index, element := range request.Args {
			keys_index[element] = index
		}
		it := db.Connector.NewIterator(read_options)
		defer it.Close()
		it.Seek([]byte(start))
		for ; it.Valid(); it.Next() {
			if bytes.Compare(it.Key(), []byte(end)) > 1 {
				break
			}
			if index, present := keys_index[string(it.Key())]; present {
				data[index] = string(it.Value())
			}
		}
	}
	response = NewSuccessResponse(data)
	db.Connector.ReleaseSnapshot(snapshot)
	return response, nil
}

func templateMGet(b *testing.B, numKeys int, fun func(*Db, *Request) (*Response, error)) {
	f := func(db_store *DbStore, db *Db) {
		fillNKeys(db, b.N)
		get := make([]string, numKeys)
		for i := 0; i < numKeys; i++ {
			get[i] = fmt.Sprintf("key_%i", i)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fun(db, &Request{Args: get})
		}
	}
	TemplateDbTest(b.Fatalf, f)
}

func Benchmark10MGet(b *testing.B) {
	templateMGet(b, 10, MGet)
}

//func Benchmark10MGetOldVersion(b *testing.B) {
//templateMGet(b, 10, OldMGet)
//}

//func BenchmarkFullScanMGet(b *testing.B) {
//templateMGet(b, b.N, MGet)
//}

//func BenchmarkFullScanMGetOldVersion(b *testing.B) {
//templateMGet(b, b.N, OldMGet)
//}
