package server

import (
	"fmt"
	"go/ast"
	"reflect"
	"sync/atomic"
)

type service struct {
	name     string //struct name
	typ      reflect.Type
	receiver reflect.Value
	method   map[string]*methodType
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Pointer {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}
func (m *methodType) newReplyv() reflect.Value {
	// reply must be a pointer type
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

func (s *service) registerMethod() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
	}
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.receiver, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// 原因在于无法确定用户传入的s.rcvr类型为结构体还是为指针，如果用户传入的为指针的话，直接采用s.typ.Name()输出的为空字符串
//
// log.Println(" struct name: " + reflect.TypeOf(foo).Name())//输出Foo
//
// log.Println("pointer name: " + reflect.TypeOf(&foo).Name())//输出空串
//
// 因此需要采用reflect.Indirect(s.rcvr)方法，提取实例对象再获取名称
func newService(receive interface{}) *service {
	s := new(service)
	s.receiver = reflect.ValueOf(receive)
	s.name = reflect.Indirect(s.receiver).Type().Name()
	s.typ = reflect.TypeOf(receive)
	if !ast.IsExported(s.name) {
		panic(fmt.Sprintf("rpc server:%s is nit a valid service name", s.name))
	}
	s.registerMethod()
	return s
}
