package server

import (
	"flag"
	goconfig "github.com/msbranco/goconfig"
	"reflect"
)

func loadConfigFromFile(path string, obj interface{}, section string) error {
	iniConfig, err := goconfig.ReadConfigFile(path)
	if err != nil {
		return err
	}

	config := reflect.ValueOf(obj).Elem()
	configType := config.Type()

	for i := 0; i < config.NumField(); i++ {
		structField := config.Field(i)
		fieldTag := configType.Field(i).Tag.Get("ini")
		switch {
		case structField.Type().Kind() == reflect.Bool:
			configValue, err := iniConfig.GetBool(section, fieldTag)
			if err == nil {
				structField.SetBool(configValue)
			}
		case structField.Type().Kind() == reflect.String:
			configValue, err := iniConfig.GetString(section, fieldTag)
			if err == nil {
				structField.SetString(configValue)
			}
		case structField.Type().Kind() == reflect.Int:
			configValue, err := iniConfig.GetInt64(section, fieldTag)
			if err == nil {
				structField.SetInt(configValue)
			}
		}
	}
	return nil
}

func setFlag(fs *flag.FlagSet, obj interface{}) error {
	config := reflect.ValueOf(obj).Elem()
	configType := config.Type()
	for i := 0; i < config.NumField(); i++ {
		structField := config.Field(i)
		shortFlag := configType.Field(i).Tag.Get("short")
		description := configType.Field(i).Tag.Get("description")
		if shortFlag == "" {
			continue
		}
		switch {
		case structField.Type().Kind() == reflect.Bool:
			v := structField.Addr().Interface().(*bool)
			fs.BoolVar(v, shortFlag, structField.Bool(), description)
		case structField.Type().Kind() == reflect.String:
			v := structField.Addr().Interface().(*string)
			fs.StringVar(v, shortFlag, structField.String(), description)
		case structField.Type().Kind() == reflect.Int:
			v := structField.Addr().Interface().(*int)
			fs.IntVar(v, shortFlag, int(structField.Int()), description)
		}
	}
	return nil
}
