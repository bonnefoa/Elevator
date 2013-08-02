package elevator

import (
	"bytes"
	"github.com/ugorji/go/codec"
	"os"
	"strings"
	"os/signal"
	"syscall"
	"log"
)

func DirExists(path string) (bool, error) {
	file_info, err := os.Stat(path)
	if err != nil {
		return false, err
	} // if file doesn't exists, throws here

	return file_info.IsDir(), nil
}

func IsFilePath(str string) bool {
	startswith_dot := strings.HasPrefix(str, ".")
	contains_slash := strings.Contains(str, "/")

	if startswith_dot == true || contains_slash == true {
		return true
	}

	return false
}

func Truncate(str string, l int) string {
	var truncated bytes.Buffer

	if len(str) > l {
		for i := 0; i < l; i++ {
			truncated.WriteString(string(str[i]))
		}
	} else {
		return str
	}

	return truncated.String()
}

func MegabytesToBytes(mb int) int {
	return mb * 1048576
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// PackInto method fulfills serializes a value
// into a msgpacked response message
func PackInto(v interface{}, buffer *bytes.Buffer) error {
	ptr_v := &v
	var handler codec.MsgpackHandle
	enc := codec.NewEncoder(buffer, &handler)
	err := enc.Encode(ptr_v)
	if err != nil {
		return err
	}
	return nil
}

// UnpackFrom method fulfills a value from a received
// serialized request message.
func UnpackFrom(v interface{}, data *bytes.Buffer) error {
	ptr_v := &v
	var handler codec.MsgpackHandle
	dec := codec.NewDecoder(data, &handler)
	err := dec.Decode(ptr_v)
	if err != nil {
		return err
	}
	return nil
}

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
