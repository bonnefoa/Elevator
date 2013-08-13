package store

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type TypeCommand int

const (
	STORE_COMMAND = iota
	DB_COMMAND
	UNKNOWN_COMMAND
)

type Request struct {
	DbUid   string
	Command string
	Args    [][]byte
	Id      [][]byte
	TypeCommand TypeCommand
}

var _ = fmt.Print

// String represents the Request as a normalized string
func (r Request) String() string {
	if len(r.Args) < 10 {
		return fmt.Sprintf("<Request uid:%s command:%s args:%s>",
			r.DbUid, r.Command, r.Args)
	}
	return fmt.Sprintf("<Request uid:%s command:%s args:%s...>",
		r.DbUid, r.Command, r.Args[0:10])
}

func GetTypeRequest(command string) TypeCommand {
	if _, foundStoreCommand := storeCommands[command]; foundStoreCommand {
		return STORE_COMMAND
	}
	if _, foundDbCommand := databaseComands[command]; foundDbCommand {
		return DB_COMMAND
	}
	return UNKNOWN_COMMAND
}

func RequestFromByte(req []byte) (*Request, error) {
	words := bytes.Split(req, []byte(" "))
	if len(words) == 0 {
		return nil, EmptyCommand(req)
	}
	cmd := string(bytes.ToUpper(words[0]))
	typeCommand := GetTypeRequest(cmd)
	if typeCommand == UNKNOWN_COMMAND {
		return nil, UnknownCommand(cmd)
	}
	args := words[1:]
	return &Request{Command: cmd, Args: args}, nil
}

func (request *Request) SendRequest(socket *zmq.Socket) {
	buffer := bytes.Buffer{}
	PackInto(request, &buffer)
	socket.SendMultipart([][]byte{buffer.Bytes()}, 0)
}

func PartsToRequest(parts [][]byte) (*Request, error) {
	request := &Request{}
	id := parts[0:2]
	rawMsg := parts[2]
	var msg *bytes.Buffer = bytes.NewBuffer(rawMsg)
	// Deserialize request message and fulfill request
	// obj with it's content
	UnpackFrom(request, msg)
	request.Id = id

	request.TypeCommand = GetTypeRequest(request.Command)
	if request.TypeCommand == UNKNOWN_COMMAND {
		return request, UnknownCommand(request.Command)
	}
	return request, nil
}
