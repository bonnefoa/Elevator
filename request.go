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
	cmd := strings.ToUpper(words[0])
	_, exist_store := store_commands[cmd]
	_, exist_db := database_commands[cmd]
	if exist_store || exist_db {
		args := words[1:]
		return &Request{Command: cmd, Args: args}, nil
	}
	return nil, errors.New(
		fmt.Sprintf("Unknown command %s", req))
}

func (request *Request) SendRequest(socket *zmq.Socket) {
	var buffer bytes.Buffer
	PackInto(request, &buffer)
	socket.SendMultipart([][]byte{buffer.Bytes()}, 0)
}
