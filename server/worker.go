package server

import (
	"code.google.com/p/goprotobuf/proto"
	zmq "github.com/bonnefoa/go-zeromq"
	"github.com/golang/glog"
	store "github.com/oleiade/Elevator/store"
	"sync"
)

// worker accepts incoming request from server and
// relay them to the dbstore
type worker struct {
	*store.DbStore
	repSocket *zmq.Socket
	*zmq.Context
	partsChannel chan *zmq.MessageMultipart
}

// sendResponse sends request result throught the
// push socket of the worker to the router
func (w *worker) sendResponse(response *Response) error {
	data, err := proto.Marshal(response)
	if err != nil {
		return err
	}
	if glog.V(8) {
		glog.Infof("Sending response %v", data)
	}
	return w.repSocket.Send(data, 0)
}

func (w *worker) processRequest(msg *zmq.MessagePart) *Response {
	request := &store.Request{}
	err := proto.Unmarshal(msg.Data, request)
	msg.Close()
	if err != nil {
		glog.Info("Worker: Error on message reading %s", err)
		return responseFromError(err)
	}
	res, err := w.DbStore.HandleRequest(request)
	if err != nil {
		return responseFromError(err)
	}
	status := Response_SUCCESS
	errMsg := ""
	response := &Response{
		Status:   &status,
		Data:     res,
		ErrorMsg: &errMsg,
	}
    return response
}

// destroyWorker clean up
func (w *worker) destroyWorker(wg *sync.WaitGroup) {
	w.repSocket.Close()
	wg.Done()
}

// startWorker bind response socket and
// wait for router requests
func (w worker) startWorker(wg *sync.WaitGroup) {
	var err error
	w.repSocket, err = createAndConnectSocket(w.Context, zmq.Rep, responseInproc)
	if err != nil {
		glog.Error("Error on starting worker ", err)
	}
	defer w.destroyWorker(wg)
	for {
		msg, err := w.repSocket.Recv(0)
		if err != nil {
			if glog.V(2) {
				glog.Warning("Error on receive ", err)
			}
			return
		}
        response := w.processRequest(msg)
        err = w.sendResponse(response)
        if err != nil {
            glog.Warning("Error when sending response", err)
        }
	}
}
