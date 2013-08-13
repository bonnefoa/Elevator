package server

import (
	"bytes"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	store "github.com/oleiade/Elevator/store"
	"log"
)

const monitorInproc = "inproc://close"
const responseInproc = "inproc://response"

type ServerState struct {
	*zmq.Context
	receiveSocket  *zmq.Socket
	responseSocket *zmq.Socket
	dbStore        *store.DbStore
	*Config
	recvChannel chan [][]byte
	exitChannel chan bool
}

// Creates and binds the zmq socket for the server
// to listen on
func (s *ServerState) initializeServer() (err error) {
	s.Context, err = zmq.NewContext()
	if err != nil {
		return
	}
	s.receiveSocket, err = s.NewSocket(zmq.ROUTER)
	if err != nil {
		return
	}
	err = s.receiveSocket.Bind(s.Endpoint)
	if err != nil {
		return
	}
	s.dbStore, err = store.InitializeDbStore(s.StoreConfig)
	if err != nil {
		return
	}
	s.responseSocket, err = s.NewSocket(zmq.PULL)
	if err != nil {
		return
	}
	err = s.responseSocket.Bind(responseInproc)
	return
}

func (s *ServerState) closeServer() {
	l4g.Info("Unmounting databases")
	s.dbStore.UnmountAll()
	l4g.Info("Closing receive socket")
	s.receiveSocket.Close()
	l4g.Info("Closing response socket")
	s.responseSocket.Close()
	l4g.Info("Closing context")
	s.Context.Close()
	l4g.Info("Closing receive and exit channel")
	close(s.recvChannel)
	close(s.exitChannel)
}

func ReceiveResponse(socket *zmq.Socket) *Response {
	response := &Response{}
	parts, err := socket.RecvMultipart(0)
	if err != nil {
		l4g.Warn("Error on response receive ", err)
	}
	store.UnpackFrom(response, bytes.NewBuffer(parts[0]))
	return response
}

func (s *ServerState) LoopPolling() {
	// Poll for events on the zmq socket
	// and send incoming requests in the recv channel
	for {
		pollers := zmq.PollItems{
			zmq.PollItem{Socket: s.receiveSocket, Events: zmq.POLLIN},
			zmq.PollItem{Socket: s.responseSocket, Events: zmq.POLLIN},
		}
		_, err := zmq.Poll(pollers, -1)
		if err != nil {
			l4g.Info("Exiting loop polling")
			return
		}
		if pollers[0].REvents&zmq.POLLIN > 0 {
			parts, err := s.receiveSocket.RecvMultipart(0)
			if err != nil {
				l4g.Warn("Error on receive ", err)
				continue
			}
			s.recvChannel <- parts
		}
		if pollers[1].REvents&zmq.POLLIN > 0 {
			parts, err := s.responseSocket.RecvMultipart(0)
			if err != nil {
				l4g.Warn("Error on response receive ", err)
				continue
			}
			err = s.receiveSocket.SendMultipart(parts, 0)
			if err != nil {
				l4g.Warn("Error on response send ", err)
			}
		}
	}
}

func ListenAndServe(config *Config, exitChannel chan bool) {
	l4g.Info(fmt.Sprintf("Elevator started on %s", config.Endpoint))
	serverState := &ServerState{Config: config, recvChannel: make(chan [][]byte, 100),
		exitChannel:exitChannel}
	err := serverState.initializeServer()
	if err != nil {
		log.Fatal(err)
	}
	workerExitChannel := make(chan bool, 0)
	worker := Worker{serverState.dbStore, nil, serverState.Context,
		serverState.recvChannel, workerExitChannel}
	for i := 0; i < 5; i++ {
		go worker.StartWorker()
	}
	go serverState.LoopPolling()
	<-exitChannel
	l4g.Info("Exiting server")
	// Closing workers
	close(workerExitChannel)
	// Closing sockets and context
	serverState.closeServer()
}
