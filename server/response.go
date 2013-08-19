package server

import (
	store "github.com/oleiade/Elevator/store"
)

func responseFromError(err error) *Response {
	status := errorToStatusCode(err)
    msg := err.Error()
    return &Response{ Status: &status, ErrorMsg:&msg }
}

func errorToStatusCode(err error) Response_Status {
	if err == nil {
		return Response_SUCCESS
	}
	switch err.(type) {
	case store.KeyError:
		return Response_KEY_ERROR
	case store.ValueError:
		return Response_VALUE_ERROR
	case store.DatabaseError:
		return Response_DATABASE_ERROR
	case store.RequestError:
		return Response_REQUEST_ERROR
	case store.UnknownCommand:
		return Response_UNKOWN_COMMAND
	}
	return Response_UNKNOWN_ERROR
}
