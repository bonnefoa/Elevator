package elevator

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
)

type Response struct {
	Status   int
	Err_code int
	Err_msg  string
	Data     []string
}

// String represents the Response as a normalized string
func (r *Response) String() string {
	return fmt.Sprintf("<Response status:%d err_code:%d err_msg:%s data:%s",
		r.Status, r.Err_code, r.Err_msg, r.Data)
}

// NewResponse returns a pointer to a brand new allocated Response
func NewResponse(status int, err_code int, err_msg string, data []string) *Response {
	return &Response{
		Status:   status,
		Err_code: err_code,
		Err_msg:  err_msg,
		Data:     data,
	}
}

// NewSuccessResponse returns a pointer to a brand
// new allocated succesfull Response
func NewSuccessResponse(data []string) *Response {
	return &Response{
		Status: SUCCESS_STATUS,
		Data:   data,
	}
}

// NewFailureResponse returns a pointer to a brand
// new allocated failure Response
func NewFailureResponse(err_code int, err_msg string) *Response {
	return &Response{
		Status:   FAILURE_STATUS,
		Err_code: err_code,
		Err_msg:  err_msg,
	}
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
		if err != nil && err.Error() == eint_error {
			continue
		}
	}
	UnpackFrom(response, bytes.NewBuffer(parts[0]))
	return response
}
