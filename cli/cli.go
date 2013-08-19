package main

import (
	"errors"
	"flag"
	"fmt"
	linenoise "github.com/GeertJohan/go.linenoise"
	zmq "github.com/bonnefoa/go-zeromq"
	"github.com/oleiade/Elevator/server"
	"github.com/oleiade/Elevator/store"
	"os"
	"strings"
    "strconv"
)

var historyFile = "$HOME/.elevator_history"

type cliState struct {
	endpoint  *string
	currentDb *string
	socket    *zmq.Socket
}

var cmdLists = make([]string, 0)

func init() {
	for _, v := range store.StoreRequest_Command_name {
		cmdLists = append(cmdLists, v)
	}
	for _, v := range store.DbRequest_Command_name {
		cmdLists = append(cmdLists, v)
	}
	historyFile = os.ExpandEnv(historyFile)
}

func writeHelp() {
	fmt.Println("HELP write this message")
	fmt.Println("QUIT stop the program")
	fmt.Println()

	fmt.Println("Store commands")
	fmt.Println("CREATE  <dbname>")
	fmt.Println("DROP    <dbname>")
	fmt.Println("LIST")

	fmt.Println()
	fmt.Println("Db commands")
	fmt.Println("USE <dbname>. All following db command will be made on the given dbname")
	fmt.Println("GET <key>")
	fmt.Println("RANGE <first key> <last key>")
	fmt.Println("MGET <key1> <key2> ...")
	fmt.Println("PUT <key> <value>")
	fmt.Println("BATCH bput <key> <value> bdel <key>")
}

func storeRequestFromString(storeCmd store.StoreRequest_Command, split []string) (*store.StoreRequest, error) {
	r := store.StoreRequest{Command: &storeCmd}
	switch storeCmd {
	case store.StoreRequest_CREATE:
		fallthrough
	case store.StoreRequest_DROP:
		if len(split) < 1 {
			return nil, errors.New("Expected a dbname parameter")
		}
		r.DbName = &split[0]
	}
	return &r, nil
}

func (c *cliState) dbRequestFromString(dbCmd store.DbRequest_Command, split []string) (*store.DbRequest, error) {
    r := &store.DbRequest{Command: &dbCmd, DbName:c.currentDb}
	switch dbCmd {
	case store.DbRequest_GET:
		if len(split) < 1 {
			return nil, errors.New("Expected a key")
		}
        r.Get = &store.GetRequest{Key:[]byte(split[0])}
    case store.DbRequest_PUT:
		if len(split) < 2 {
			return nil, errors.New("Expected a key and a value")
		}
        r.Put = &store.PutRequest{Key:[]byte(split[0]), Value:[]byte(split[1])}
    case store.DbRequest_DELETE:
		if len(split) < 1 {
			return nil, errors.New("Expected a key")
		}
        r.Delete = &store.DeleteRequest{Key:[]byte(split[0])}
    case store.DbRequest_RANGE:
		if len(split) < 2 {
			return nil, errors.New("Expected a start key and an end key")
		}
        r.Range = &store.RangeRequest{Start:[]byte(split[0]), End:[]byte(split[1])}
    case store.DbRequest_SLICE:
		if len(split) < 2 {
			return nil, errors.New("Expected a start key and a limit")
		}
        limit, err := strconv.Atoi(split[1])
        if err != nil {
            return nil, err
        }
        limit32 := int32(limit)
        r.Slice = &store.SliceRequest{Start:[]byte(split[0]), Limit:&limit32}
    case store.DbRequest_MGET:
        r.Mget = &store.MgetRequest{Keys:store.ToBytes(split...)}
	}
	return r, nil
}

func (c *cliState) requestFromString(line string) (r *store.Request, err error) {
	r = &store.Request{}
	splitted := strings.Split(line, " ")
	if len(splitted) == 0 {
		return nil, errors.New("No command")
	}
	firstWord := strings.ToUpper(splitted[0])
	storeCmd, storeFound := store.StoreRequest_Command_value[firstWord]
	dbCmd, dbFound := store.DbRequest_Command_value[firstWord]
	if !storeFound && !dbFound {
		return nil, fmt.Errorf("Unknown command %s", firstWord)
	}
	if storeFound {
		r.StoreRequest, err = storeRequestFromString(store.StoreRequest_Command(storeCmd), splitted[1:])
		rcmd := store.Request_STORE
		r.Command = &rcmd
	}
	if dbFound {
		r.DbRequest, err = c.dbRequestFromString(store.DbRequest_Command(dbCmd), splitted[1:])
		rcmd := store.Request_DB
		r.Command = &rcmd
	}
	return r, err
}

func (c *cliState) parseRequestFromLine(line string) (*store.Request, error) {
	upperString := strings.ToUpper(line)
	if upperString == "HELP" {
		writeHelp()
		return nil, nil
	}
	if strings.HasPrefix(upperString, "USE") {
		err := c.useDb(line)
		return nil, err
	}
	request, err := c.requestFromString(line)
	if err != nil {
		return nil, err
	}
	return request, err
}

func completionHandler(str string) []string {
	res := []string{}
	upperString := strings.ToUpper(str)
	for _, cmd := range cmdLists {
		if strings.HasPrefix(cmd, upperString) {
			res = append(res, cmd)
		}
	}
	return res
}

func (c *cliState) useDb(str string) error {
	splitted := strings.Split(str, " ")
	if len(splitted) < 2 {
		return errors.New("Missing dbname parameter")
	}
	c.currentDb = &splitted[1]
	return nil
}

func (c *cliState) getLine() (string, bool) {
    prompt := "> "
    if c.currentDb != nil {
        prompt = fmt.Sprintf("(%s)> ", *c.currentDb)
    }
    line, err := linenoise.Line(prompt)
    line = strings.TrimSpace(line)
    if err != nil {
        if err == linenoise.KillSignalError {
            return "", true
        }
        fmt.Printf("Unexpected error : %s\n", err)
    }
    isQuit := strings.ToUpper(line) == "QUIT"
    return line, isQuit
}

func (c *cliState) loop() {
	for {
        // Fetch line from cli
        line, isQuit := c.getLine()
        if isQuit {
            return
        }

		request, err := c.parseRequestFromLine(line)
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}
		linenoise.AddHistory(line)
        if request == nil {
            continue
        }
		err = server.SendRequest(request, c.socket)
		if err != nil {
			fmt.Printf("Error on send request %q\n", err)
			continue
		}
		response, err := server.ReceiveResponse(c.socket)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(response)
		}
	}
}

func main() {
	state := &cliState{}
	state.endpoint = flag.String("e",
		server.DefaultEndpoint, "Endpoint of elevator")
	flag.Parse()

	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.Req)
	socket.Connect(*state.endpoint)
	state.socket = socket

	linenoise.LoadHistory(historyFile)
	linenoise.SetCompletionHandler(completionHandler)
	state.loop()
	err := linenoise.SaveHistory(historyFile)
	if err != nil {
		fmt.Println(err)
	}
}
