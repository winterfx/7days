package geerpc

import (
	"encoding/json"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

// | Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
// | <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|
// 由于不同的编码方式，可能存在粘包等问题，RPC消息阶段粘包问题（gob/json实现了"RPC消息拆包"）
// 你说的问题应该是server端解析Option的时候可能会破坏后面RPC消息的完整性，当客户端消息发送过快服务端消息积压时（例：Option|Header|Body|Header|Body），服务端使用json解析Option，json.Decode()调用conn.read()读取数据到内部的缓冲区（例：Option|Header），此时后续的RPC消息就不完整了(Body|Header|Body)。
// 示例代码中客户端简单的使用time.sleep()方式隔离协议交换阶段与RPC消息阶段，减少这种问题发生的可能。
type Server struct {
}

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

type request struct {
	h           *codec.Header
	argv, reply reflect.Value
}

func (s *Server) handleRequest(c codec.Codec, r *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	/// TODO, should call registered rpc methods to get the right reply
	//	// day 1, just print argv and send a hello message
	defer wg.Done()
	r.reply = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", r.h.Seq))
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
	req.argv = reflect.New(reflect.TypeOf(""))
	if err := c.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
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
