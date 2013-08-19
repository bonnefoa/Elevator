package server

import (
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
