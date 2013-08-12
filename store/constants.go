package store

type DbMountedStatus int
type DbCommand string
type StoreCommand string

const (
	DB_STATUS_MOUNTED = DbMountedStatus(1)
	DB_STATUS_UNMOUNTED = DbMountedStatus(0)
)

// Command codes
const (
	DB_GET     = "GET"
	DB_PUT     = "PUT"
	DB_DELETE  = "DELETE"
	DB_RANGE   = "RANGE"
	DB_SLICE   = "SLICE"
	DB_BATCH   = "BATCH"
	DB_MGET    = "MGET"
	DB_PING    = "PING"

	DB_CONNECT = "DBCONNECT"
	DB_MOUNT   = "DBMOUNT"
	DB_UMOUNT  = "DBUMOUNT"
	DB_CREATE  = "DBCREATE"
	DB_DROP    = "DBDROP"
	DB_LIST    = "DBLIST"
	DB_REPAIR  = "DBREPAIR"
)

// batches signals
const (
	SIGNAL_BATCH_PUT    = "BPUT"
	SIGNAL_BATCH_DELETE = "BDEL"
)
