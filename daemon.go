package elevator

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func createPidFile(pidfile string) error {
	if pidString, err := ioutil.ReadFile(pidfile); err == nil {
		pid, err := strconv.Atoi(string(pidString))
		if err == nil {
			if _, err := os.Stat(fmt.Sprintf("/proc/%d/", pid)); err == nil {
				return fmt.Errorf("pid file found, ensure elevator is not running or delete %s", pidfile)
			}
		}
	}

	file, err := os.Create(pidfile)
	if err != nil {
		log.Println(err)
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	return err
}

func removePidFile(pidfile string) {
	if err := os.Remove(pidfile); err != nil {
		log.Printf("Error removing %s: %s", pidfile, err)
	}
}

func Daemon(config *Config, exitChannel chan bool) {
	if err := createPidFile(config.Pidfile); err != nil {
		log.Fatal(err)
	}
	defer removePidFile(config.Pidfile)
	ListenAndServe(config, exitChannel)
}
