package server

import (
	"fmt"
	store "github.com/oleiade/Elevator/store"
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

func ResponseFromError(id [][]byte, err error) *Response {
	status := ErrorToStatusCode(err)
	return &Response {
		Status: status,
		ErrMsg: err.Error(),
		Id:id,
	}
}


type ResponseStatus int

const (
	SUCCESS        = ResponseStatus(0)
	TYPE_ERROR     = ResponseStatus(1)
	KEY_ERROR      = ResponseStatus(2)
	VALUE_ERROR    = ResponseStatus(3)
	INDEX_ERROR    = ResponseStatus(4)
	RUNTIME_ERROR  = ResponseStatus(5)
	OS_ERROR       = ResponseStatus(6)
	DATABASE_ERROR = ResponseStatus(7)
	SIGNAL_ERROR   = ResponseStatus(8)
	REQUEST_ERROR  = ResponseStatus(9)
	UNKNOWN_ERROR  = ResponseStatus(10)
	UNKOWN_COMMAND  = ResponseStatus(11)
)

func ErrorToStatusCode(err error) ResponseStatus {
	if err == nil {
		return SUCCESS
	}
	switch err.(type) {
	case store.KeyError:
		return KEY_ERROR
	case store.ValueError:
		return VALUE_ERROR
	case store.DatabaseError:
		return DATABASE_ERROR
	case store.RequestError:
		return REQUEST_ERROR
	case store.UnknownCommand:
		return UNKOWN_COMMAND
	}
	return UNKNOWN_ERROR
}
