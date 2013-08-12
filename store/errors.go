package store

import (
	"fmt"
)

type KeyError []byte
type ValueError error
type DatabaseError error
type NoSuchDbError string
type NoSuchDbuidError string
type UnknownCommand string
type EmptyCommand string
type RequestError error

func (k KeyError) Error() string {
	return fmt.Sprintf("Key %s does not exists", k)
}

func (k NoSuchDbError) Error() string {
	return fmt.Sprintf("No such db %s", k)
}

func (k NoSuchDbuidError) Error() string {
	return fmt.Sprintf("No such db uid %s", k)
}

func (c EmptyCommand) Error() string {
	return fmt.Sprintf("Empty command in request %s", c)
}

func (c UnknownCommand) Error() string {
	return fmt.Sprintf("Unkown command %s", c)
}
