package main

import (
	"bufio"
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	elevator "github.com/oleiade/Elevator"
	"os"
)

type CliState struct {
	endpoint   *string
	currentUid *string
	socket     *zmq.Socket
}

func main() {
	conf := &CliState{}
	conf.endpoint = flag.String("e", elevator.DEFAULT_ENDPOINT, "Endpoint of elevator")
	flag.Parse()

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	socket.Connect(*conf.endpoint)
	conf.socket = socket

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		l := scanner.Text()
		req, err := elevator.RequestFromString(l)
		if err != nil {
			fmt.Print(err)
			fmt.Print("\n\n")
			continue
		}
		req.SendRequest(socket)
		response := elevator.ReceiveResponse(socket)
		fmt.Print(response)
		fmt.Print("\n\n")
	}
}
