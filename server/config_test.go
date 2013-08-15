package server

import (
	"flag"
	"reflect"
	"testing"
)

func TestConfigFromPath(t *testing.T) {
	expected := newConfig()
	c, _ := ConfFromFile("conf/elevator.conf")
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("\nExpected \n%q\ngot \n%q", expected, c)
	}
}

func TestConfigFromCommandLine(t *testing.T) {
	fs := flag.NewFlagSet("TOTo", flag.ContinueOnError)
	confFile := fs.String("c",
		"GA",
		"Specifies config file path")
	config := newConfig()
	SetFlag(fs, config.ServerConfig)

	err := fs.Parse([]string{"-c", "toto", "-d"})
	t.Log(fs.Args())
	t.Log(err)
	if *confFile != "toto" {
		t.Fatalf("Expected toto, got %s", *confFile)
	}
	if !config.Daemon {
		t.Fatalf("Expected Daemon to be true %s", config)
	}

}
