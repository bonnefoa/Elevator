package elevator

import (
	"fmt"
	"bytes"
	"errors"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	"log"
)

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
	socket.Bind(endpoint)
	return socket, context, nil
}

// handleRequest deserializes the input msgpack request,
// processes it and ensures it is forwarded to the client.
func handleRequest(client_socket *ClientSocket, raw_msg []byte, db_store *DbStore) {
	var request 	*Request = new(Request)
	var msg 		*bytes.Buffer = bytes.NewBuffer(raw_msg)

	// Deserialize request message and fulfill request
	// obj with it's content
	UnpackFrom(request, msg)
	request.Source = client_socket
	l4g.Debug(func() string { return request.String() })

	if request.DbUid != "" {
		if db, ok := db_store.Container[request.DbUid]; ok {
			if db.Status == DB_STATUS_UNMOUNTED {
				db.Mount()
			}
			db.Channel <- request
		}
	} else if len(request.Args) > 0 {
		go func() {
			response, err := store_commands[request.Command](db_store, request)
			if err != nil {
				l4g.Error(err)
			}
			forwardResponse(response, request)
		}()
	} else {
		response := NewFailureResponse(REQUEST_ERROR, "Invalid arguments")
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
	error := errors.New(fmt.Sprintf("Unknown command %s", request.Command))
	l4g.Error(error)

	return nil, error
}

// forwardResponse takes a request-response pair as input and
// sends the response to the request client.
func forwardResponse(response *Response, request *Request) error {
	l4g.Debug(func() string { return response.String() })

	var response_buf bytes.Buffer
	var socket *zmq.Socket = &request.Source.Socket

	PackInto(response, &response_buf)

	parts := request.Source.Id
	parts = append(parts, response_buf.Bytes())
	err := socket.SendMultipart(parts, 0)
	if err != nil {
		return err
	}

	return nil
}

func PollChannel(socket *zmq.Socket, pollChan chan [][]byte) {
	// Poll for events on the zmq socket
	// and send incoming requests in the poll channel
	for {
		// build zmq poller
		pollers := zmq.PollItems{
			zmq.PollItem{Socket: socket, Events: zmq.POLLIN},
		}
		zmq.Poll(pollers, -1)
                parts, _ := pollers[0].Socket.RecvMultipart(0)
                pollChan <- parts
	}
}

func ListenAndServe(config *Config, exitSignal chan bool) {
	l4g.Info(fmt.Sprintf("Elevator started on %s", config.Core.Endpoint))

	// Build server zmq socket
	socket, context, err := buildServerSocket(config.Core.Endpoint)

	if err != nil {
		log.Fatal(err)
	}
	// Load database store
	db_store := NewDbStore(config)
	defer func() {
		db_store.UnmountAll()
		socket.Close()
		context.Close()
		exitSignal <- true
	}()

	err = db_store.Load()
	if err != nil {
		err = db_store.Add("default")
		if err != nil {
			log.Fatal(err)
		}
	}

	pollChan := make(chan [][]byte)
	go PollChannel(socket, pollChan)

	for {
		select {
		case parts := <-pollChan:
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
