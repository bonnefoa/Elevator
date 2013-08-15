package store

import (
	"fmt"
)

// KeyError happens when unknown key is requested
type KeyError []byte
// ValueError happens when a put operation has failed
type ValueError error
// DatabaseError happens on unexpected error
type DatabaseError error
// NoSuchDbError happens when a requested dbname does not exists
type NoSuchDbError string
// NoSuchDbUIDError happens when a requested db UID does not exists
type NoSuchDbUIDError string
// UnknownCommand happens when request command is unknown
type UnknownCommand string
// EmptyCommand happens when request command is empty
type EmptyCommand string
// RequestError happens when request is invalid
type RequestError error

func (k KeyError) Error() string {
	return fmt.Sprintf("Key %q does not exists", []byte(k))
}

func (k NoSuchDbError) Error() string {
	return fmt.Sprintf("No such db %q", string(k))
}

func (k NoSuchDbUIDError) Error() string {
	return fmt.Sprintf("No such db uid %q", string(k))
}

func (c EmptyCommand) Error() string {
	return fmt.Sprintf("Empty command in request %q", string(c))
}

func (c UnknownCommand) Error() string {
	return fmt.Sprintf("Unkown command %q", string(c))
}
