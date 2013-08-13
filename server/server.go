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

func PollChannel(socket *zmq.Socket, pollChan chan [][]byte, exitSignal chan bool) {
	// Poll for events on the zmq socket
	// and send incoming requests in the poll channel
	for {
		// build zmq poller
		pollers := zmq.PollItems{
			zmq.PollItem{Socket: socket, Events: zmq.POLLIN},
		}
		_, err := zmq.Poll(pollers, time.Second)
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
    dbStore := store.NewDbStore(config.StoreConfig)
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

    worker := Worker { dbStore, nil, context, pollChan, exitSignal }
    for i:=0; i < 10; i++ {
        worker.StartWorker()
    }

	for {
		select {
		case <-exitSignal:
			l4g.Info("Exiting server")
			return
		}
	}
}
