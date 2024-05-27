package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

// | Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
// | <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
// 由于不同的编码方式，可能存在粘包等问题，RPC消息阶段粘包问题（gob/json实现了"RPC消息拆包"）
// 你说的问题应该是server端解析Option的时候可能会破坏后面RPC消息的完整性，当客户端消息发送过快服务端消息积压时（例：Option|Header|Body|Header|Body），服务端使用json解析Option，json.Decode()调用conn.read()读取数据到内部的缓冲区（例：Option|Header），此时后续的RPC消息就不完整了(Body|Header|Body)。
// 示例代码中客户端简单的使用time.sleep()方式隔离协议交换阶段与RPC消息阶段，减少这种问题发生的可能。
type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (s *Server) ServerConn(coon net.Conn) {
	defer func() {
		_ = coon.Close()
	}()
	var opt codec.Option
	if err := json.NewDecoder(coon).Decode(&opt); err != nil {
		return
	}
	if !opt.IsGeeRpc() {
		return
	}
	if f, ok := codec.NewCodecFuncMap[opt.CodecType]; !ok {
		return
	} else {
		s.serverCodec(f(coon))
	}
}

func (s *Server) handleRequest(c codec.Codec, r *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	/// TODO, should call registered rpc methods to get the right reply
	//	// day 1, just print argv and send a hello message
	defer wg.Done()
	err := r.svc.call(r.mtype, r.argv, r.reply)
	if err != nil {
		r.h.Error = err.Error()
		s.sendResponse(c, r.h, invalidRequest, sending)
	}
	s.sendResponse(c, r.h, r.reply.Interface(), sending)

}
func (s *Server) serverCodec(c codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)
	for {
		wg.Add(1)
		//var h *codec.Header
		r, err := s.readRequest(c)
		if err != nil {
			if r == nil {
				break
			}
			r.h.Error = err.Error()
			s.sendResponse(c, r.h, invalidRequest, sending)
			continue
		}
		go s.handleRequest(c, r, sending, wg)
	}
	wg.Wait()
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt codec.Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if !opt.IsGeeRpc() {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

// invalidRequest is a placeholder for response argv when error occurs

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
func (s *Server) readRequest(c codec.Codec) (*request, error) {
	var h *codec.Header
	if err := c.ReadHeader(h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			fmt.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	req := &request{
		h: h,
	}
	var err error
	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.reply = req.mtype.newReplyv()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err := c.ReadBody(argvi); err != nil {
		fmt.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}
func (s *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(h, body); err != nil {
		fmt.Println("rpc server: write response error:", err)
	}
}
func (s *Server) Accept(lis net.Listener) {
	for {
		coon, err := lis.Accept()
		if err != nil {
			return
		}
		go s.ServerConn(coon)
	}
}
func (s *Server) Register(rcvr interface{}) error {
	sc := newService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(sc.name, sc); dup {
		return errors.New("rpc:service already defined:" + sc.name)
	}
	return nil
}

func Register(receiver interface{}) error {
	return DefaultServer.Register(receiver)
}
