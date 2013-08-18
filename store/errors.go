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
// RelativePathError happens when a dbname is a relative filepath
type RelativePathError string
// NoSuchPathError happens when an absolute dbname does not exists
type NoSuchPathError string
// DatabaseExistsError happens when creating an already present database
type DatabaseExistsError string
// MissingDbNameError happens when a request needs dbname parameter
type MissingDbNameError string
// DbAlreadyMounted happens when a trying to mount an already mounted db
type DbAlreadyMounted string
// DbAlreadyUnmounted happens when a trying to unmount an already unmounted db
type DbAlreadyUnmounted string

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

func (c RelativePathError) Error() string {
	return fmt.Sprintf("Creating database from relative path not allowed (was %q)", string(c))
}

func (c NoSuchPathError) Error() string {
	return fmt.Sprintf("%s does not exists", string(c))
}

func (c DatabaseExistsError) Error() string {
	return fmt.Sprintf("Database %s already exists", string(c))
}

func (c MissingDbNameError) Error() string {
	return fmt.Sprintf("DbName parameter is needed for command %s", string(c))
}

func (c DbAlreadyMounted) Error() string {
	return fmt.Sprintf("Database %s already mounted", string(c))
}

func (c DbAlreadyUnmounted) Error() string {
	return fmt.Sprintf("Database %s already unmounted", string(c))
}
