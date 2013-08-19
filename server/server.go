package server

import (
	"fmt"
	zmq "github.com/bonnefoa/go-zeromq"
	"github.com/golang/glog"
	store "github.com/oleiade/Elevator/store"
	"log"
	"sync"
)

const responseInproc = "inproc://response"

type serverState struct {
	*zmq.Context
	frontendSocket *zmq.Socket
	backendSocket  *zmq.Socket
	dbStore        *store.DbStore
	*config
	exitChannel chan bool
}

// Creates and binds the zmq socket for the server
// to listen on
func (s *serverState) initializeServer() (err error) {
	s.Context, err = zmq.NewContext()
	s.frontendSocket, err = createAndBindSocket(s.Context, zmq.Router, s.Endpoint)
	if err != nil {
		return
	}
	s.backendSocket, err = createAndBindSocket(s.Context, zmq.Dealer, responseInproc)
	if err != nil {
		return
	}
	s.dbStore, err = store.InitializeDbStore(s.StoreConfig)
	if err != nil {
		return
	}
	return
}

func (s *serverState) closeServer() {
	if glog.V(4) {
		glog.Info("Unmounting databases")
	}
	s.dbStore.UnmountAll()
	if glog.V(4) {
		glog.Info("Closing pool socket")
	}
	if glog.V(4) {
		glog.Info("Closing receive socket")
	}
	s.frontendSocket.Close()
	if glog.V(4) {
		glog.Info("Closing backend socket")
	}
	s.backendSocket.Close()
	if glog.V(4) {
		glog.Info("Closing context")
	}
	s.Context.Destroy()
	if glog.V(4) {
		glog.Info("Closing receive and exit channel")
	}
}

func (s *serverState) LoopPolling() (err error) {
	// Poll for events on the zmq socket
	// and send incoming requests in the backend
	pollFrontend := &zmq.PollItem{Socket: s.frontendSocket, Events: zmq.Pollin}
	pollBackend := &zmq.PollItem{Socket: s.backendSocket, Events: zmq.Pollin}
	pollers := zmq.PollItems{pollFrontend, pollBackend}
	for {
		_, err = pollers.Poll(-1)
		if err != nil {
            if glog.V(2) {
                glog.Info("Exiting loop polling")
            }
			return err
		}
		if pollers[0].REvents&zmq.Pollin > 0 {
			msg, err := s.frontendSocket.RecvMultipart(0)
			if err != nil {
				glog.Warning("Error on receiving ", err)
				continue
			}
			s.backendSocket.SendMultipart(msg.Data, zmq.DontWait)
		}
		if pollers[1].REvents&zmq.Pollin > 0 {
			msg, err := s.backendSocket.RecvMultipart(0)
			if err != nil {
				glog.Warning("Error on receiving from backend ", err)
				continue
			}
			err = s.frontendSocket.SendMultipart(msg.Data, zmq.DontWait)
			msg.Close()
			if err != nil {
				glog.Warning("Error on sending to frontend ", err)
				continue
			}
		}
	}
}

// ListenAndServe starts listening socket, initialize worker
// and loop until an exit signal is received
func ListenAndServe(config *config, exitChannel chan bool) {
	serverState := &serverState{config: config,
		exitChannel: exitChannel}
	err := serverState.initializeServer()
	if err != nil {
		log.Fatal(err)
	}
	worker := worker{DbStore: serverState.dbStore, Context: serverState.Context}
	wg := &sync.WaitGroup{}
	for i := 0; i < config.NumWorkers; i++ {
		go worker.startWorker(wg)
	}
	wg.Add(config.NumWorkers)
	go serverState.LoopPolling()
	glog.Info(fmt.Sprintf("Elevator started on %s", config.Endpoint))
	<-exitChannel
	serverState.closeServer()
	wg.Wait()
	close(exitChannel)
}
