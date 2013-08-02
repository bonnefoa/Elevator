package elevator

import (
	"testing"
    "reflect"
)

func TestConfigFromPath(t *testing.T) {
    expected := NewConfig()
	c := &Config{
		&CoreConfig{},
		&StorageEngineConfig{},
        &LogConfiguration{},
	}
	c.FromFile("conf/elevator.conf")
	if !reflect.DeepEqual(c, expected) {
		t.Fatalf("\nExpected \n%q\ngot \n%q", expected, c)
	}
}
