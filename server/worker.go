package server

import (
	"bytes"
	zmq "github.com/bonnefoa/go-zeromq"
	store "github.com/oleiade/Elevator/store"
	"github.com/golang/glog"
    "sync"
)

// worker accepts incoming request from server and
// relay them to the dbstore
type worker struct {
	*store.DbStore
	*zmq.Socket
	*zmq.Context
	partsChannel chan [][]byte
	exitChannel  chan bool
}

// sendResponse sends request result throught the
// push socket of the worker to the router
func (w *worker) sendResponse(response *Response) {
	var responseBuf bytes.Buffer
	store.PackInto(response, &responseBuf)
	parts := response.id
	parts = append(parts, responseBuf.Bytes())
	w.Socket.SendMultipart(parts, 0)
}

func (w *worker) sendErrorResponse(id [][]byte, err error) {
	response := responseFromError(id, err)
	w.sendResponse(response)
}

func (w *worker) startResponseSocket() error {
	socket, err := w.NewSocket(zmq.Push)
	if err != nil {
		return err
	}
	err = socket.Connect(responseInproc)
	if err != nil {
		return err
	}
	w.Socket = socket
	return nil
}

func (w *worker) processRequest(parts [][]byte) {
	request, err := store.PartsToRequest(parts)
	if err != nil {
		glog.Info("Worker: Error on message reading %s", err)
		w.sendErrorResponse(request.ID, err)
		return
	}
	res, err := w.DbStore.HandleRequest(request)
	if err != nil {
		w.sendErrorResponse(request.ID, err)
	}
	response := &Response{
		Status: Success,
		Data:   res,
		id:     request.ID,
	}
	w.sendResponse(response)
}

// destroyWorker clean up
func (w *worker) destroyWorker(wg *sync.WaitGroup) {
	w.Socket.Close()
    wg.Done()
}

// startWorker bind response socket and
// wait for router requests
func (w worker) startWorker(wg *sync.WaitGroup) {
	w.startResponseSocket()
	defer w.destroyWorker(wg)
	for {
		select {
		case parts := <-w.partsChannel:
			if len(parts) < 3 {
				continue
			}
			w.processRequest(parts)
		case <-w.exitChannel:
			glog.Info("Received exit signal, destroying worker")
			return
		}
	}
}
