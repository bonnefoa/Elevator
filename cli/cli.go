package main

import (
	"flag"
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	zmq "github.com/alecthomas/gozmq"
	elevator "github.com/oleiade/Elevator"
	"strings"
)

type cliState struct {
	endpoint   *string
	currentDb  *string
	currentUid *string
	socket     *zmq.Socket
}

func writeHelp() {
	fmt.Println("HELP write this message")
	fmt.Println("QUIT stop the program")

	fmt.Println("DBCREATE  <dbname>")
	fmt.Println("DBDROP    <dbname>")
	fmt.Println("DBCONNECT <dbname>")
	fmt.Println("DBLIST")
	fmt.Println("GET <key>")
	fmt.Println("RANGE <first key> <last key>")
	fmt.Println("MGET <key1> <key2> ...")
	fmt.Println("PUT <key> <value>")
	fmt.Println("BATCH BPUT <key> <value> BDEL <key>")
}

func processRequest(line string, state *cliState, socket *zmq.Socket) *elevator.Request {
	request, err := elevator.RequestFromString(line)
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

func completionHandler(str string) []string {
	res := []string{}
	upperString := strings.ToUpper(str)
	for _, cmd := range elevator.CommandsList {
		if strings.HasPrefix(cmd, upperString) {
			res = append(res, cmd)
		}
	}
	return res
}

func cliLoop(state *cliState) {
	for {
		prompt := "> "
		if state.currentUid != nil {
			prompt = fmt.Sprintf("(%s)> ", *state.currentDb)
		}
		line, err := linenoise.Line(prompt)
		upperString := strings.ToUpper(line)
		if err != nil {
			if err == linenoise.KillSignalError {
				return
			}
			fmt.Println("Unexpected error : %s", err)
		}
		if upperString == "HELP" {
			writeHelp()
			continue
		}
		if upperString == "QUIT" {
			return
		}
		request := processRequest(line, state, state.socket)
		if request == nil {
			continue
		}
		response := elevator.ReceiveResponse(state.socket)
		if request.Command == elevator.DB_CONNECT &&
			response.Status == elevator.SUCCESS_STATUS {
			state.currentDb = &request.Args[0]
			state.currentUid = &response.Data[0]
		}
		linenoise.AddHistory(line)
		fmt.Print(response)
		fmt.Print("\n")
	}
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

	linenoise.SetCompletionHandler(completionHandler)
	cliLoop(state)
	linenoise.SaveHistory("~/.elevator_history")
}
