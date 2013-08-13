package server

import (
	"bytes"
	zmq "github.com/alecthomas/gozmq"
	l4g "github.com/alecthomas/log4go"
	store "github.com/oleiade/Elevator/store"
)

type Worker struct {
	*store.DbStore
	*zmq.Socket
	*zmq.Context
	partsChannel chan [][]byte
	exitChannel  chan bool
}

func (w *Worker) sendResponse(response *Response) {
	var responseBuf bytes.Buffer
	store.PackInto(response, &responseBuf)
	parts := response.Id
	parts = append(parts, responseBuf.Bytes())
	w.Socket.SendMultipart(parts, 0)
}

func (w *Worker) sendErrorResponse(id [][]byte, err error) {
	response := ResponseFromError(id, err)
	w.sendResponse(response)
}

func (w *Worker) startResponseSocket() error {
	socket, err := w.NewSocket(zmq.PUSH)
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

func (w *Worker) processRequest(parts [][]byte) {
	request, err := store.PartsToRequest(parts)
	if err != nil {
		l4g.Info("Error on message reading %s", err)
		w.sendErrorResponse(request.Id, err)
		return
	}
	l4g.Debug(request.String())
	res, err := w.HandleRequest(request)
	if err != nil {
		w.sendErrorResponse(request.Id, err)
	}
	response := &Response{
		Status: SUCCESS,
		Data:   res,
		Id:     request.Id,
	}
	w.sendResponse(response)
}

func (w *Worker) DestroyWorker() {
	l4g.Info("Destroying worker")
	w.Socket.Close()
}

func (w Worker) StartWorker() {
	w.startResponseSocket()
	defer w.DestroyWorker()
	for {
		select {
		case parts := <-w.partsChannel:
			if len(parts) < 3 {
				continue
			}
			w.processRequest(parts)
		case <-w.exitChannel:
			l4g.Debug("Received exit signal, destroying worker")
			return
		}
	}
}
