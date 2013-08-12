package server

import (
	"reflect"
	goconfig "github.com/msbranco/goconfig"
	"flag"
)

func LoadConfigFromFile(path string, obj interface{}, section string) error {
	ini_config, err := goconfig.ReadConfigFile(path)
	if err != nil {
		return err
	}

	config := reflect.ValueOf(obj).Elem()
	config_type := config.Type()

	for i := 0; i < config.NumField(); i++ {
		struct_field := config.Field(i)
		field_tag := config_type.Field(i).Tag.Get("ini")
		switch {
		case struct_field.Type().Kind() == reflect.Bool:
			config_value, err := ini_config.GetBool(section, field_tag)
			if err == nil {
				struct_field.SetBool(config_value)
			}
		case struct_field.Type().Kind() == reflect.String:
			config_value, err := ini_config.GetString(section, field_tag)
			if err == nil {
				struct_field.SetString(config_value)
			}
		case struct_field.Type().Kind() == reflect.Int:
			config_value, err := ini_config.GetInt64(section, field_tag)
			if err == nil {
				struct_field.SetInt(config_value)
			}
		}
	}
	return nil
}

func SetFlag(fs *flag.FlagSet, obj interface{}) error {
	config := reflect.ValueOf(obj).Elem()
	config_type := config.Type()
	for i := 0; i < config.NumField(); i++ {
		struct_field := config.Field(i)
		short_flag := config_type.Field(i).Tag.Get("short")
		description := config_type.Field(i).Tag.Get("description")
		if short_flag == "" { continue }
		switch {
		case struct_field.Type().Kind() == reflect.Bool:
			v := struct_field.Addr().Interface().(*bool)
			fs.BoolVar(v, short_flag, struct_field.Bool(), description)
		case struct_field.Type().Kind() == reflect.String:
			v := struct_field.Addr().Interface().(*string)
			fs.StringVar(v, short_flag, struct_field.String(), description)
		case struct_field.Type().Kind() == reflect.Int:
			v := struct_field.Addr().Interface().(*int)
			fs.IntVar(v, short_flag, int(struct_field.Int()), description)
		}
	}
	return nil
}
