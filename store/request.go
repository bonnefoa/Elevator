package store

import (
	"bytes"
	"fmt"
	zmq "github.com/bonnefoa/go-zeromq"
	"github.com/golang/glog"
)

// RequestType specificies if the requests should be handled by the dbstore or a specific db
type requestType int

const (
	typeStore = iota
	typeDb
	typeUnkown
)

// Request allows to query a specific db if a DbUID is given
// You can also query the dbstore to create or drop a db
type Request struct {
	DbUID   string
	Command string
	Args    [][]byte
	ID      [][]byte
	requestType requestType
}

// String represents the Request as a normalized string
func (r Request) String() string {
	if len(r.Args) < 10 {
		return fmt.Sprintf("<Request uid:%s command:%s args:%s>",
			r.DbUID, r.Command, r.Args)
	}
	return fmt.Sprintf("<Request uid:%s command:%s args:%s...(%d)>",
		r.DbUID, r.Command, r.Args[0:10], len(r.Args))
}

func getRequestType(command string) requestType {
	if _, foundStoreCommand := storeCommands[command]; foundStoreCommand {
		return typeStore
	}
	if _, foundDbCommand := databaseComands[command]; foundDbCommand {
		return typeDb
	}
	return typeUnkown
}

// SendRequest pack request and send it via the given zero mq socket
func (r *Request) SendRequest(socket *zmq.Socket) {
    if glog.V(1) {
        glog.Info("Sending request ", r)
    }
	buffer := bytes.Buffer{}
	PackInto(r, &buffer)
	socket.SendMultipart([][]byte{buffer.Bytes()}, 0)
}

// PartsToRequest parse parts from a message receveived from a
// zmq router socket (with first parts being socket id) to a request
func PartsToRequest(parts [][]byte) (*Request, error) {
	request := &Request{}
	id := parts[0:2]
	rawMsg := parts[2]
	msg := bytes.NewBuffer(rawMsg)
	// Deserialize request message and fulfill request
	// obj with it's content
    if glog.V(1) {
        glog.Info("Unpacking message %s", msg)
    }
	UnpackFrom(request, msg)
	request.ID = id

	request.requestType = getRequestType(request.Command)
	if request.requestType == typeUnkown {
		return request, UnknownCommand(request.Command)
	}
	return request, nil
}
