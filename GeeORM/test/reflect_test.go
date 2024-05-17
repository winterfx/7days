package test

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type Config struct {
	Name    string `json:"sever-name"`
	IP      string `json:"server-ip"`
	URL     string `json:"server-url"`
	Timeout string `json:"timeout"`
}

func readConfig() *Config {
	config := Config{}
	typ := reflect.TypeOf(config)
	s := typ.Name()
	fmt.Println(s)
	value := reflect.Indirect(reflect.ValueOf(&config))
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if v, ok := f.Tag.Lookup("json"); ok {
			key := fmt.Sprintf("CONFIG_%s", strings.ReplaceAll(strings.ToUpper(v), "-", "_"))
			if env, exist := os.LookupEnv(key); exist {
				value.FieldByName(f.Name).Set(reflect.ValueOf(env))
			}
		}
	}
	return &config
}
func TestReflect(t *testing.T) {
	readConfig()
}
