package store

import (
	"github.com/ugorji/go/codec"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// ToBytes convert strings passed as argument to a slice of bytes
func ToBytes(args...string) [][]byte {
	res := make([][]byte, len(args))
	for i, v := range args {
		res[i] = []byte(v)
	}
	return res
}

func dirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	} // if file doesn't exists, throws here

	return fileInfo.IsDir(), nil
}

func isFilePath(str string) bool {
	startsWithDot := strings.HasPrefix(str, ".")
	containsSlash := strings.Contains(str, "/")

	if startsWithDot == true || containsSlash == true {
		return true
	}

	return false
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

var msgpackHandler = codec.MsgpackHandle{}

// PackInto method fulfills serializes a value
// into a msgpacked response message
func PackInto(v interface{}, buffer io.Writer) error {
	prtV := &v
	enc := codec.NewEncoder(buffer, &msgpackHandler)
	err := enc.Encode(prtV)
	if err != nil {
		return err
	}
	return nil
}

// UnpackFrom method fulfills a value from a received
// serialized request message.
func UnpackFrom(v interface{}, data io.Reader) error {
	prtV := &v
	dec := codec.NewDecoder(data, &msgpackHandler)
	err := dec.Decode(prtV)
	if err != nil {
		return err
	}
	return nil
}

// SetupExitChannel creates a channel which received a value on
// SIGTERM
func SetupExitChannel() chan bool {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, os.Signal(syscall.SIGTERM))
	exitChannel := make(chan bool)
	go func() {
		sig := <-c
		log.Printf("Received signal '%v', exiting\n", sig)
		exitChannel <- true
	}()
	return exitChannel
}
