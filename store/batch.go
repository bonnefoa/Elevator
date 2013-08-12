package store

import (
	leveldb "github.com/jmhodges/levigo"
	"fmt"
)

type BatchOperations []BatchOperation

type PutOperation struct {
    key []byte
    value []byte
}

type DeleteOperation struct {
    key []byte
}

type BatchOperation interface {
	ExecuteBatch(*leveldb.WriteBatch)
}

func (p PutOperation) ExecuteBatch(wb *leveldb.WriteBatch) {
	wb.Put(p.key, p.value)
}

func (p DeleteOperation) ExecuteBatch(wb *leveldb.WriteBatch) {
	wb.Delete(p.key)
}

// BatchOperationsFromRequestArgs builds a BatchOperations from
// a string slice resprensenting a sequence of batch operations
func BatchOperationsFromRequestArgs(args [][]byte) (BatchOperations, error) {
	var ops BatchOperations
	var op BatchOperation
	for i:=0; i < len(args); i++ {
		if string(args[i]) == SIGNAL_BATCH_PUT {
			if len(args) < i + 2 {
				return ops, fmt.Errorf("Not enough arguments after %q", args[i:])
			}
			op = PutOperation{args[i+1], args[i+2]}
			i += 2
		} else {
			if len(args) < i + 1 {
				return ops, fmt.Errorf("Not enough arguments after %q", args[i:])
			}
			op = DeleteOperation{args[i+1]}
			i += 1
		}
		ops = append(ops, op)
	}
	return ops, nil
}
