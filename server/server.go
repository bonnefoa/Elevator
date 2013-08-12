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

func ReceiveResponse(socket *zmq.Socket) *Response {
	var response *Response
	var parts [][]byte
	var err error
	for {
		parts, err = socket.RecvMultipart(0)
		if err == nil {
			break
		}
	}
	store.UnpackFrom(response, bytes.NewBuffer(parts[0]))
	return response
}

func sendResponse(socket *zmq.Socket, response *Response) {
	var response_buf bytes.Buffer
	store.PackInto(response, &response_buf)
}

func sendErrorResponse(socket *zmq.Socket, id [][]byte, err error) {
	response := ResponseFromError(id, err)
	sendResponse(socket, response)
}

// handleRequest deserializes the input msgpack request,
// processes it and ensures it is forwarded to the client.
func handleRequest(dbStore *store.DbStore) {
	// Deserialize request message and fulfill request
	// obj with it's content
	for {
		request, err := store.PartsToRequest(parts)
		if err != nil {
			l4g.Info("Error on message reading %s", err)
			sendErrorResponse(request.Id, err)
			return
		}
		l4g.Debug(func() string { return request.String() })
		res, err := dbStore.HandleRequest(request)
		if err != nil {
			sendErrorResponse(request.Id, err)
		}
		response := &Response {
			Status:SUCCESS,
			Data:res,
			Id:request.Id,
		}
		sendResponse(response)
	}
}

// forwardResponse takes a request-response pair as input and
// sends the response to the request client.
func forwardResponse(response *Response, request *Request) error {
	l4g.Debug(func() string { return response.String() })

	var err error
	var response_buf bytes.Buffer
	var socket *zmq.Socket = &request.source.Socket


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
