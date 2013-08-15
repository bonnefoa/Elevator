package store

// DbMountedStatus identifies if a connector to the leveldb database
// is opened
type DbMountedStatus int

const (
	statusUnmounted DbMountedStatus = iota
	statusMounted
)

// Command codes
const (
	DbGet     = "GET"
	DbPut     = "PUT"
	DbDelete  = "DELETE"
	DbRange   = "RANGE"
	DbSlice   = "SLICE"
	DbBatch   = "BATCH"
	DbMget    = "MGET"
	DbPing    = "PING"

	DbConnect = "DBCONNECT"
	DbMount   = "DBMOUNT"
	DbUmount  = "DBUMOUNT"
	DbCreate  = "DBCREATE"
	DbDrop    = "DBDROP"
	DbList    = "DBLIST"
	DbRepair  = "DBREPAIR"
)

// batches signals
const (
	SignalBatchPut    = "BPUT"
	SignalBatchDelete = "BDEL"
)
