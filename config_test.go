package elevator

import (
	"flag"
	"reflect"
	"testing"
)

func TestConfigFromPath(t *testing.T) {
	expected := NewConfig()
	c, _ := ConfFromFile("conf/elevator.conf")
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("\nExpected \n%q\ngot \n%q", expected, c)
	}
}

func TestConfigFromCommandLine(t *testing.T) {
	fs := flag.NewFlagSet("TOTo", flag.ContinueOnError)
	conf_file := fs.String("c",
		"GA",
		"Specifies config file path")
	config := NewConfig()
	SetFlag(fs, config.CoreConfig)

	err := fs.Parse([]string{"-c", "toto", "-d"})
	t.Log(fs.Args())
	t.Log(err)
	if *conf_file != "toto" {
		t.Fatalf("Expected toto, got %s", *conf_file)
	}
	if !config.Daemon {
		t.Fatalf("Expected Daemon to be true", config)
	}
}
