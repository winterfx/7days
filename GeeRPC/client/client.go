package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"net"
	"sync"
)

var ErrShutdown = errors.New("connection is shut down")

type Client struct {
	cc       codec.Codec
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	opt      *codec.Option
	seq      uint64
	pending  map[uint64]*Call //线程不安全，需要mu
	closing  bool
	shutdown bool
}

func (c *Client) RegisterCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}
func (c *Client) receive() {
	var err error
	//loop receive response
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			// call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了。
			err = c.cc.ReadBody(nil)
		case len(h.Error) != 0:
			call.Error = fmt.Errorf(h.Error)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body" + err.Error())
			}
			call.done()
		}
	}
}
func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.RegisterCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

func NewClient(conn net.Conn, opt *codec.Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		fmt.Println("rpc client: codec error:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		_ = conn.Close()
		return nil, err
	}
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      f(conn),
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client, nil
}
func parseOptions(options ...*codec.Option) (*codec.Option, error) {
	opt := options[0]
	opt.MagicNumber = codec.DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = codec.DefaultOption.CodecType
	}
	return opt, nil

}
func Dial(network, address string, opts ...*codec.Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}
