package main

import (
	"bufio"
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	elevator "github.com/oleiade/Elevator"
	"os"
)

type cliState struct {
	endpoint   *string
	currentDb  *string
	currentUid *string
	socket     *zmq.Socket
}

func printHelp() {
	fmt.Printf(
		`DBCREATE  <dbname>
DBDROP    <dbname>
DBCONNECT <dbname>
DBLIST
GET <key>
PUT <key> <value>
`)
}

func processRequest(state *cliState, scanner *bufio.Scanner, socket *zmq.Socket) *elevator.Request {
	scanner.Scan()
	l := scanner.Text()
	if l == "help" {
		printHelp()
		return nil
	}
	request, err := elevator.RequestFromString(l)
	if err != nil {
		fmt.Print(err)
		fmt.Print("\n")
		return nil
	}
	if state.currentUid != nil {
		request.DbUid = *state.currentUid
	}
	request.SendRequest(socket)
	return request
}

func main() {
	state := &cliState{}
	state.endpoint = flag.String("e",
		elevator.DEFAULT_ENDPOINT, "Endpoint of elevator")
	flag.Parse()

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REQ)
	socket.Connect(*state.endpoint)
	state.socket = socket

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if state.currentUid == nil {
			fmt.Print("$> ")
		} else {
			fmt.Printf("(%s) $> ", *state.currentDb)
		}
		request := processRequest(state, scanner, socket)
		if request == nil {
			continue
		}
		response := elevator.ReceiveResponse(socket)
		if request.Command == elevator.DB_CONNECT &&
			response.Status == elevator.SUCCESS_STATUS {
			state.currentDb = &request.Args[0]
			state.currentUid = &response.Data[0]
		}
		fmt.Print(response)
		fmt.Print("\n")
	}
}
