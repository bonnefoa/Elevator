package server

import (
	"bytes"
	"fmt"
	zmq "github.com/bonnefoa/go-zeromq"
	store "github.com/oleiade/Elevator/store"
	"log"
	"github.com/golang/glog"
)

const monitorInproc = "inproc://close"
const responseInproc = "inproc://response"

type serverState struct {
	*zmq.Context
	receiveSocket  *zmq.Socket
	responseSocket *zmq.Socket
	dbStore        *store.DbStore
	*config
	recvChannel chan [][]byte
	exitChannel chan bool
}

// Creates and binds the zmq socket for the server
// to listen on
func (s *serverState) initializeServer() (err error) {
	s.Context, err = zmq.NewContext()
	if err != nil {
		return
	}
	s.receiveSocket, err = s.NewSocket(zmq.Router)
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
	s.responseSocket, err = s.NewSocket(zmq.Pull)
	if err != nil {
		return
	}
	err = s.responseSocket.Bind(responseInproc)
	return
}

func (s *serverState) closeServer() {
	glog.Info("Unmounting databases")
	s.dbStore.UnmountAll()
	glog.Info("Closing receive socket")
	s.receiveSocket.Close()
	glog.Info("Closing response socket")
	s.responseSocket.Close()
	glog.Info("Closing context")
	s.Context.Destroy()
	glog.Info("Closing receive and exit channel")
	close(s.recvChannel)
	close(s.exitChannel)
}

// ReceiveResponse fetch an incoming response from the socket
// The message is expected to be single frame
func ReceiveResponse(socket *zmq.Socket) *Response {
	response := &Response{}
	parts, err := socket.RecvMultipart(0)
	if err != nil {
		glog.Warning("Error on response receive ", err)
	}
	store.UnpackFrom(response, bytes.NewBuffer(parts.Data[0]))
	parts.Close()
	return response
}

func (s *serverState) LoopPolling() {
	// Poll for events on the zmq socket
	// and send incoming requests in the recv channel
    pollReceive := &zmq.PollItem{Socket: s.receiveSocket, Events: zmq.Pollin, REvents:0}
    pollResponse := &zmq.PollItem{Socket: s.responseSocket, Events: zmq.Pollin, REvents:0}
	pollers := zmq.PollItems{ pollReceive, pollResponse }
	for {
		_, err := pollers.Poll(-1)
		if err != nil {
			glog.Info("Exiting loop polling")
			return
		}
		if pollers[0].REvents&zmq.Pollin > 0 {
			parts, err := s.receiveSocket.RecvMultipart(0)
			if err != nil {
				glog.Warning("Error on receive ", err)
				continue
			}
			s.recvChannel <- parts.Data
			parts.Close()
		}
		if pollers[1].REvents&zmq.Pollin > 0 {
			parts, err := s.responseSocket.RecvMultipart(0)
			if err != nil {
				glog.Warning("Error on response receive ", err)
				continue
			}
			err = s.receiveSocket.SendMultipart(parts.Data, 0)
			parts.Close()
			if err != nil {
				glog.Warning("Error on response send ", err)
			}
		}
	}
}

// ListenAndServe starts listening socket, initialize worker
// and loop until an exit signal is received
func ListenAndServe(config *config, exitChannel chan bool) {
	glog.Info(fmt.Sprintf("Elevator started on %s", config.Endpoint))
	serverState := &serverState{config: config, recvChannel: make(chan [][]byte, 100),
		exitChannel: exitChannel}
	err := serverState.initializeServer()
	if err != nil {
		log.Fatal(err)
	}
	workerExitChannel := make(chan bool, 0)
	worker := worker{serverState.dbStore, nil, serverState.Context,
		serverState.recvChannel, workerExitChannel}
	for i := 0; i < 5; i++ {
		go worker.startWorker()
	}
	go serverState.LoopPolling()
	<-exitChannel
	glog.Info("Exiting server")
	// Closing workers
	close(workerExitChannel)
	// Closing sockets and context
	serverState.closeServer()
}
