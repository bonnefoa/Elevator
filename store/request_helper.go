package store

func newStoreRequest(dbName string, cmd StoreRequest_Command) StoreRequest {
	return StoreRequest{&dbName, &cmd, nil}
}

// NewStoreRequest creates a store request
func NewStoreRequest(dbName string, storeCmd StoreRequest_Command) Request {
	cmd := Request_STORE
	r := &StoreRequest{&dbName, &storeCmd, nil}
	return Request{Command: &cmd, StoreRequest: r}
}

// NewPutRequest creates a db request for a put
func NewPutRequest(dbName string, key []byte, val []byte) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_PUT
	r.Command = &cmd
	r.Put = &PutRequest{key, val, nil}
	return r
}

// NewGetRequest creates a db request for a get
func NewGetRequest(dbName string, key []byte) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_GET
	r.Command = &cmd
	r.Get = &GetRequest{key, nil}
	return r
}

// NewDeleteRequest creates a db request for a delete
func NewDeleteRequest(dbName string, key []byte) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_DELETE
	r.Command = &cmd
	r.Delete = &DeleteRequest{key, nil}
	return r
}

// NewMgetRequest creates a db request for a multiple get
func NewMgetRequest(dbName string, keys ...[]byte) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_MGET
	r.Command = &cmd
	r.Mget = &MgetRequest{keys, nil}
	return r
}

// NewRangeRequest creates a db request for a range request
func NewRangeRequest(dbName string, start []byte, end []byte) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_RANGE
	r.Command = &cmd
	r.Range = &RangeRequest{start, end, nil}
	return r
}

// NewSliceRequest creates a db request for a slice request
func NewSliceRequest(dbName string, start []byte, limit int32) DbRequest {
	r := DbRequest{DbName: &dbName}
	cmd := DbRequest_SLICE
	r.Command = &cmd
	r.Slice = &SliceRequest{start, &limit, nil}
	return r
}

// NewBatchRequest creates a db request for a batch request
func NewBatchRequest(dbName string, putKeys [][]byte, putsValues [][]byte,
	deleteKeys [][]byte) DbRequest {
	r := DbRequest{DbName: &dbName}
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
	r.Batch = &BatchRequest{batchPuts, batchDeletes, nil}
	return r
}
