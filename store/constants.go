package store

// DbMountedStatus identifies if a connector to the leveldb database
// is opened
type DbMountedStatus int

const (
	statusUnmounted DbMountedStatus = iota
	statusMounted
)
