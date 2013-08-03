package elevator

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"strings"
)

type Request struct {
	DbUid   string
	Command string
	Args    []string
	Source  *ClientSocket `msgpack:"-"`
}

var DatabaseRequest = map[string]bool{
	DB_GET:    true,
	DB_MGET:   true,
	DB_PUT:    true,
	DB_DELETE: true,
	DB_RANGE:  true,
	DB_SLICE:  true,
	DB_BATCH:  true,
}

var DbstoreRequests = map[string]bool{
	DB_CREATE:  true,
	DB_DROP:    true,
	DB_CONNECT: true,
	DB_MOUNT:   true,
	DB_UMOUNT:  true,
	DB_LIST:    true,
}

// String represents the Request as a normalized string
func (r *Request) String() string {
	return fmt.Sprintf("<Request uid:%s command:%s args:%s>",
		r.DbUid, r.Command, r.Args)
}

func RequestFromString(req string) (*Request, error) {
	words := strings.Split(req, " ")
	if len(words) == 0 {
		return nil, errors.New("Empty request")
	}
	cmd := words[0]
	if _, exist := DatabaseRequest[cmd]; exist {
		if len(words) < 3 {
			return nil, errors.New(
				fmt.Sprintf("Expected CMD DBUID ARGS, got %s", req))
		}
		dbuid := words[1]
		args := words[2:]
		return &Request{DbUid: dbuid, Command: cmd, Args: args}, nil
	}
	if _, exist := DbstoreRequests[cmd]; exist {
		if len(words) < 2 {
			return nil, errors.New(
				fmt.Sprintf("Expected CMD ARGS, got %s", req))
		}
		args := words[1:]
		return &Request{Command: cmd, Args: args}, nil
	}
	return nil, errors.New(
		fmt.Sprintf("Unknown command %s", req))
}

// NewRequest returns a pointer to a brand new allocated Request
func NewRequest(command string, args []string) *Request {
	return &Request{
		Command: command,
		Args:    args,
	}
}

func (request *Request) SendRequest(socket *zmq.Socket) {
	var buffer bytes.Buffer
	PackInto(request, &buffer)
	socket.SendMultipart([][]byte{buffer.Bytes()}, 0)

}
