package store

import (
	leveldb "github.com/jmhodges/levigo"
	"fmt"
)

type batchOperations []batchOperation

type putOperation struct {
    key []byte
    value []byte
}

type deleteOperation struct {
    key []byte
}

type batchOperation interface {
	executeBatch(*leveldb.WriteBatch)
}

func (p putOperation) executeBatch(wb *leveldb.WriteBatch) {
	wb.Put(p.key, p.value)
}

func (p deleteOperation) executeBatch(wb *leveldb.WriteBatch) {
	wb.Delete(p.key)
}

// batchOperationsFromRequestArgs builds a batchOperations from
// a string slice resprensenting a sequence of batch operations
func batchOperationsFromRequestArgs(args [][]byte) (batchOperations, error) {
	var ops batchOperations
	var op batchOperation
	for i:=0; i < len(args); i++ {
		if string(args[i]) == SignalBatchPut {
			if len(args) < i + 2 {
				return ops, fmt.Errorf("Not enough arguments after %q", args[i:])
			}
			op = putOperation{args[i+1], args[i+2]}
			i += 2
		} else if string(args[i]) == SignalBatchDelete {
			if len(args) < i + 1 {
				return ops, fmt.Errorf("Not enough arguments after %q", args[i:])
			}
			op = deleteOperation{args[i+1]}
			i += 1
		} else {
			return ops, fmt.Errorf("Unknown operator at %d (%s)", i, args[i])
		}
		ops = append(ops, op)
	}
	return ops, nil
}
