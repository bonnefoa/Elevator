package elevator

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	"log"
	"time"
)

var eint_error = "interrupted system call"

type ClientSocket struct {
	Id     [][]byte
	Socket zmq.Socket
}

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
func handleRequest(client_socket *ClientSocket, raw_msg []byte, db_store *DbStore) {
	var request *Request = new(Request)
	var msg *bytes.Buffer = bytes.NewBuffer(raw_msg)

	// Deserialize request message and fulfill request
	// obj with it's content
	UnpackFrom(request, msg)
	request.source = client_socket
	l4g.Debug(func() string { return request.String() })

	_, found_db := database_commands[request.Command]
	if found_db {
		if db, ok := db_store.Container[request.DbUid]; ok {
			if db.Status == DB_STATUS_UNMOUNTED {
				db.Mount(db_store.Config.Options)
			}
			db.Channel <- request
		} else {
			l4g.Error("Could not find %s in container %q",
				request.DbUid, db_store.Container)
			response := NewFailureResponse(KEY_ERROR,
				fmt.Sprintf("Could not find dbuid %q", request.DbUid))
			forwardResponse(response, request)
		}
	} else {
		response, err := store_commands[request.Command](db_store, request)
		if err != nil {
			l4g.Error(err)
		}
		forwardResponse(response, request)
	}
}

// processRequest executes the received request command, and returns
// the resulting response.
func processRequest(db *Db, request *Request) (*Response, error) {
	if f, ok := database_commands[request.Command]; ok {
		response, _ := f(db, request)
		return response, nil
	}
	error := fmt.Errorf("Unknown command %s", request.Command)
	l4g.Error(error)

	return nil, error
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
	db_store := NewDbStore(config)
	err = db_store.Load()
	if err != nil {
		err = db_store.Add("default")
		if err != nil {
			log.Fatal(err)
		}
	}
	pollChan := make(chan [][]byte)

	defer func() {
		db_store.UnmountAll()
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
			client_socket := ClientSocket{
				Id:     parts[0:2],
				Socket: *socket,
			}
			msg := parts[2]
			go handleRequest(&client_socket, msg, db_store)
		case <-exitSignal:
			l4g.Info("Exiting server")
			return
		}
	}
}
