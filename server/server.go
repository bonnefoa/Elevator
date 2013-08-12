package server

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	store "github.com/oleiade/Elevator/store"
	"log"
	"time"
)

var eint_error = "interrupted system call"

// Creates and binds the zmq socket for the server
// to listen on
func buildServerSocket(endpoint string) (*zmq.Socket, *zmq.Context, error) {
	context, err := zmq.NewContext()
	if err != nil {
		return nil, nil, err
	}
	socket, err := context.NewSocket(zmq.ROUTER)
	if err != nil {
		return nil, nil, err
	}
	err = socket.Bind(endpoint)
	if err != nil {
		return nil, nil, err
	}
	return socket, context, nil
}

// handleRequest deserializes the input msgpack request,
// processes it and ensures it is forwarded to the client.
func handleRequest(parts [][]byte, dbStore *store.DbStore) {
	// Deserialize request message and fulfill request
	// obj with it's content
	request := store.PartsToRequest(parts)
	l4g.Debug(func() string { return request.String() })
	_, found_db := databaseComands[request.Command]
	if !found_db {
		l4g.Error("Could not find %s in container %q", request.DbUid, dbStore.Container)
		dbError := &DbError{ KEY_ERROR, request.Args, fmt.Errorf("Could not find dbuid %q", request.DbUid)}
		response := NewFailureResponse(dbError)
		forwardResponse(response, request)
	}
	db, ok := dbStore.Container[request.DbUid]
	if !ok {
		response, err := storeCommands[request.Command](dbStore, request.Args)
		if err != nil {
			l4g.Error(err)
		}
		forwardResponse(response, request)
	}
	if db.Status == DB_STATUS_UNMOUNTED {
		db.Mount(dbStore.Config.Options)
	}
	db.Channel <- request
}

// forwardResponse takes a request-response pair as input and
// sends the response to the request client.
func forwardResponse(response *Response, request *Request) error {
	l4g.Debug(func() string { return response.String() })

	var err error
	var response_buf bytes.Buffer
	var socket *zmq.Socket = &request.source.Socket

	PackInto(response, &response_buf)

	parts := request.source.Id
	parts = append(parts, response_buf.Bytes())
	for {
		err = socket.SendMultipart(parts, 0)
		if err == nil {
			break
		}
		if err.Error() == eint_error {
			continue
		}
		l4g.Warn("Error when sending response %v", err)
		return err
	}
	return nil
}

func PollChannel(socket *zmq.Socket, pollChan chan [][]byte, exitSignal chan bool) {
	// Poll for events on the zmq socket
	// and send incoming requests in the poll channel
	for {
		// build zmq poller
		pollers := zmq.PollItems{
			zmq.PollItem{Socket: socket, Events: zmq.POLLIN},
		}
		_, err := zmq.Poll(pollers, time.Second)
		if err != nil && err.Error() == eint_error {
			continue
		}
		if err != nil {
			select {
			case <-exitSignal:
				l4g.Info("Exiting poll channel")
				return
			case <-time.After(time.Millisecond):
				l4g.Warn("Error on polling %q", err)
				continue
			}
		}
		parts, _ := pollers[0].Socket.RecvMultipart(0)
		pollChan <- parts
	}
}

func ListenAndServe(config *Config, exitSignal chan bool) {
	l4g.Info(fmt.Sprintf("Elevator started on %s", config.Endpoint))

	// Build server zmq socket
	socket, context, err := buildServerSocket(config.Endpoint)

	if err != nil {
		log.Fatal(err)
	}
	// Load database store
	dbStore := NewDbStore(config)
	err = dbStore.Load()
	if err != nil {
		err = dbStore.Add("default")
		if err != nil {
			log.Fatal(err)
		}
	}
	pollChan := make(chan [][]byte)

	defer func() {
		dbStore.UnmountAll()
		socket.Close()
		context.Close()
		close(exitSignal)
		close(pollChan)
	}()

	go PollChannel(socket, pollChan, exitSignal)

	for {
		select {
		case parts := <-pollChan:
			if len(parts) < 3 {
				continue
			}
			go handleRequest(&parts, dbStore)
		case <-exitSignal:
			l4g.Info("Exiting server")
			return
		}
	}
}
