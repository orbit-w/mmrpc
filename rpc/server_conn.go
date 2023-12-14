package rpc

import (
	"errors"
	"github.com/orbit-w/mmrpc/rpc/mmrpcs"
	"github.com/orbit-w/orbit-net/core/stream_transport"
	"io"
	"log"
	"runtime/debug"
)

type ISession interface {
	Send(pid int64, seq uint32, category int8, out []byte) error
}

type Conn struct {
	Codec
	stream stream_transport.IStreamServer
}

func NewConn(stream stream_transport.IStreamServer) {
	conn := Conn{}
	conn.stream = stream
	conn.reader()
}

func (c *Conn) Send(pid int64, seq uint32, category int8, out []byte) error {
	pack := c.Codec.encode(pid, seq, category, out)
	return c.stream.Send(pack)
}

func (c *Conn) Close() {
	_ = c.stream.Close("")
}

func (c *Conn) reader() {
	for {
		in, err := c.stream.Recv()
		if err != nil {
			switch {
			case mmrpcs.IsCancelError(err):
			case errors.Is(err, io.EOF):
			default:
				log.Println("conn read stream failed: ", err.Error())
			}
			return
		}
		if in == nil {
			log.Println("wo cao")
		}
		req, err := NewRequest(c, in)
		if err != nil {
			log.Println("[ServerConn] [reader] new request failed: ", err.Error())
			continue
		}
		c.handleRequest(req)
	}
}

func (c *Conn) handleRequest(req IRequest) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			log.Println("stack: ", string(debug.Stack()))
		}
	}()
	//TODO：need to handle user-level errors?
	_ = gRequestHandle(req)
}
