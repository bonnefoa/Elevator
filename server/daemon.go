package server

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"strconv"
)

func createPidFile(pidfile string) error {
	if pidString, err := ioutil.ReadFile(pidfile); err == nil {
		pid, err := strconv.Atoi(string(pidString))
		if err == nil {
			path := fmt.Sprintf("/proc/%d/", pid)
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("pid file found, ensure elevator is not running or delete %s", pidfile)
			}
		}
	}

	file, err := os.Create(pidfile)
	if err != nil {
		glog.Info(err)
		return err
	}

	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	return err
}

func removePidFile(pidfile string) {
	if err := os.Remove(pidfile); err != nil {
		glog.Infof("Error removing %s: %s", pidfile, err)
	}
}

// Daemon wrap ListenAndServe with the creation of a pid file
func Daemon(config *Config, exitChannel chan bool) {
	if err := createPidFile(config.Pidfile); err != nil {
		glog.Error(err)
	}
	defer removePidFile(config.Pidfile)
	ListenAndServe(config, exitChannel)
}
