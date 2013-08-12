package store

import (
	"fmt"
)

type KeyError []byte
type ValueError error
type DatabaseError error
type NoSuchDbError string
type UnknownCommand string
type RequestError error

func (k KeyError) Error() string {
	return fmt.Sprintf("Key %s does not exists", k)
}

func (k NoSuchDbError) Error() string {
	return fmt.Sprintf("No such db %s", k)
}

func (c UnknownCommand) Error() string {
	return fmt.Sprintf("Unkown command %s", c)
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
	case KeyError:
		return KEY_ERROR
	case ValueError:
		return VALUE_ERROR
	case DatabaseError:
		return DATABASE_ERROR
	case RequestError:
		return REQUEST_ERROR
	case UnknownCommand:
		return UNKOWN_COMMAND
	}
	return UNKNOWN_ERROR
}
