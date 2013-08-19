package server

import (
	"code.google.com/p/goprotobuf/proto"
	zmq "github.com/bonnefoa/go-zeromq"
	store "github.com/oleiade/Elevator/store"
)

// SendRequest marshal and transmit request throught given zmq socket
func SendRequest(r *store.Request, s *zmq.Socket) error {
	data, err := proto.Marshal(r)
	if err != nil {
		return err
	}
	return s.SendMultipart([][]byte{data}, 0)
}

// SendDbRequest marshal and transmit a db request throught given zmq socket
func SendDbRequest(dbRequest *store.DbRequest, s *zmq.Socket) error {
	cmd := store.Request_DB
	r := &store.Request{Command: &cmd, DbRequest: dbRequest}
	data, err := proto.Marshal(r)
	if err != nil {
		return err
	}
	return s.SendMultipart([][]byte{data}, 0)
}

// SendStoreRequest marshal and transmit a store request throught given zmq socket
func SendStoreRequest(storeRequest *store.StoreRequest, s *zmq.Socket) error {
	cmd := store.Request_STORE
	r := &store.Request{Command: &cmd, StoreRequest: storeRequest}
	data, err := proto.Marshal(r)
	if err != nil {
		return err
	}
	return s.SendMultipart([][]byte{data}, 0)
}

func createAndBindSocket(ctx *zmq.Context, t zmq.SocketType, addr string) (s *zmq.Socket, err error) {
	s, err = ctx.NewSocket(t)
	if err != nil {
		return
	}
	err = s.Bind(addr)
	return
}

func createAndConnectSocket(ctx *zmq.Context, t zmq.SocketType, addr string) (s *zmq.Socket, err error) {
	s, err = ctx.NewSocket(t)
	if err != nil {
		return
	}
	err = s.Connect(addr)
	return
}

// ReceiveResponse fetch an incoming response from the socket
// The message is expected to be single frame
func ReceiveResponse(socket *zmq.Socket) (*Response, error) {
	response := &Response{}
	msg, err := socket.RecvMultipart(0)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(msg.Data[0], response)
	if err != nil {
		return nil, err
	}
	msg.Close()
	return response, err
}
