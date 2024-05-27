package server

import (
	"geerpc/codec"
	"reflect"
)

type request struct {
	h           *codec.Header
	argv, reply reflect.Value
	mtype       *methodType
	svc         *service
}
