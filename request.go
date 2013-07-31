package elevator

import (
	"fmt"
)

type Request struct {
    DbUid   string
    Command string
    Args    []string
    Source  *ClientSocket `msgpack:"-"`
}

// String represents the Request as a normalized string
func (r *Request) String() string {
    return fmt.Sprintf("<Request uid:%s command:%s args:%s>",
        r.DbUid, r.Command, r.Args)
}

// NewRequest returns a pointer to a brand new allocated Request
func NewRequest(command string, args []string) *Request {
	return &Request{
		Command: command,
		Args:    args,
	}
}
