package test

import (
	"context"
	"github.com/orbit-w/mmrpc/rpc"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"testing"
	"time"
)

func Test_RPCCall(t *testing.T) {
	host := "127.0.0.1:6900"
	err := rpc.Serve(host, nil)
	assert.NoError(t, err)

	cli, err := rpc.Dial("node_00", "node_01", host)
	assert.NoError(t, err)

	for i := 0; i < 1000; i++ {
		in, err := cli.Call(context.Background(), []byte{1})
		assert.NoError(t, err)
		log.Println(in[0])
	}

	time.Sleep(time.Second * 5)
}

func TestAsyncCall(t *testing.T) {
	err := rpc.Serve("127.0.0.1:6800", nil)
	assert.NoError(t, err)

	cli, err := rpc.Dial("node_00", "node_01", "127.0.0.1:6800")
	assert.NoError(t, err)

	pid := int64(100)
	err = cli.AsyncCallC([]byte{1}, pid, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)
}

func TestBenchAsyncCall(t *testing.T) {
	err := rpc.Serve("127.0.0.1:6800", nil)
	assert.NoError(t, err)
	pid := int64(100)
	msg := []byte{3}
	cli, err := rpc.Dial("node_00", "node_01", "127.0.0.1:6800")
	assert.NoError(t, err)
	for i := 0; i < 100000; i++ {
		if err := cli.AsyncCall(msg, pid); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Second * 15)
}

func TestAsyncCallTimeout(t *testing.T) {
	err := rpc.Serve("127.0.0.1:6800", func(req rpc.IRequest) error {
		r := req.NewReader()
		data, _ := io.ReadAll(r)
		switch data[0] {
		case 3:
			return nil
		default:
			return req.Response([]byte{1})
		}
	})
	assert.NoError(t, err)

	cli, err := rpc.Dial("node_00", "node_01", "127.0.0.1:6800")
	assert.NoError(t, err)

	pid := int64(100)
	err = cli.AsyncCallC([]byte{3}, pid, func(ctx any, in []byte, err error) error {
		v := ctx.(int64)
		log.Println(v)
		log.Println("err: ", err)
		log.Println(in)
		return nil
	})
	assert.NoError(t, err)
	time.Sleep(time.Second * 15)
}

func StartPProf() {
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6999", nil))
	}()
}

func Test_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r.(string))
			log.Println("stack: ", string(debug.Stack()))
		}
	}()

	panic("invalid pointer")
}
