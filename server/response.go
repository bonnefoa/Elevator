package server

import (
	"fmt"
	store "github.com/oleiade/Elevator/store"
)

// Response represents the response send
// back by the server
type Response struct {
	Status ResponseStatus
	ErrMsg string
	Data   [][]byte
	id     [][]byte
}

// String represents the Response as a normalized string
func (r *Response) String() string {
	if r == nil {
		return "<Response nil>"
	}
	return fmt.Sprintf("<Response status:%d err_msg:%s data:%s",
		r.Status, r.ErrMsg, r.Data)
}

func responseFromError(id [][]byte, err error) *Response {
	status := errorToStatusCode(err)
	return &Response{
		Status: status,
		ErrMsg: err.Error(),
		id:     id,
	}
}

// ResponseStatus identifies status code send in the response
type ResponseStatus int

// Response status available
const (
    Success       ResponseStatus = iota
    TypeError
    KeyError
    ValueError
    IndexError
    RuntimeError
    OsError
    DatabaseError
    SignalError
    RequestError
    UnknownError
    UnkownCommand
)

func errorToStatusCode(err error) ResponseStatus {
	if err == nil {
		return Success
	}
	switch err.(type) {
	case store.KeyError:
		return KeyError
	case store.ValueError:
		return ValueError
	case store.DatabaseError:
		return DatabaseError
	case store.RequestError:
		return RequestError
	case store.UnknownCommand:
		return UnkownCommand
	}
	return UnknownError
}
