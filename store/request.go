package store

import (
	"bytes"
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type Request struct {
	DbUid   string
	Command string
	Args    [][]byte
	Id      [][]byte
}

// String represents the Request as a normalized string
func (r *Request) String() string {
	return fmt.Sprintf("<Request uid:%s command:%s args:%s>",
		r.DbUid, r.Command, r.Args)
}

func RequestFromByte(req []byte) (*Request, error) {
	words := bytes.Split(req, []byte(" "))
	if len(words) == 0 {
		return nil, errors.New("Empty request")
	}
	cmd := string(bytes.ToUpper(words[0]))
	_, exist_store := storeCommands[cmd]
	_, exist_db := databaseComands[cmd]
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

func PartsToRequest(parts [][]byte) (*Request, error) {
	var request *Request
	id := parts[0:2]
	rawMsg := parts[2]
	var msg *bytes.Buffer = bytes.NewBuffer(rawMsg)
	UnpackFrom(request, msg)
	request.Id = id

	_, foundDbCommand := databaseComands[request.Command]
	_, foundStoreCommand := storeCommands[request.Command]
	if !foundDbCommand && !foundStoreCommand {
		return nil, UnknownCommand(request.Command)
	}
	return request, nil
}
