package store

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type Response struct {
	Status   ResponseStatus
	ErrMsg   string
	Data     [][]byte
	Id	 [][]byte
}

// String represents the Response as a normalized string
func (r *Response) String() string {
	if r == nil {
		return "<Response nil>"
	}
	return fmt.Sprintf("<Response status:%d err_msg:%s data:%s",
		r.Status, r.ErrMsg, r.Data)
}

func ReceiveResponse(socket *zmq.Socket) *Response {
	response := &Response{}
	var parts [][]byte
	var err error
	for {
		parts, err = socket.RecvMultipart(0)
		if err == nil {
			break
		}
	}
	UnpackFrom(response, bytes.NewBuffer(parts[0]))
	return response
}
