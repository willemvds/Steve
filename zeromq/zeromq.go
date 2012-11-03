package zeromq

import (
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"time"
)

type Handler func([]byte)

type zeromq struct {
	context     zmq.Context
	socket      zmq.Socket
	hndChannels []chan []byte
}

func (z *zeromq) listen(addr string) error {
	var err error
	z.context, err = zmq.NewContext()
	if err != nil {
		return errors.New(fmt.Sprintf("zeromq context: %s", err))
	}
	z.socket, err = z.context.NewSocket(zmq.REP)
	if err != nil {
		return errors.New(fmt.Sprintf("zeromq socket: %s", err))
	}
	err = z.socket.Bind(addr)
	if err != nil {
		return errors.New(fmt.Sprintf("zeromq bind: %s", err))
	}
	return nil
}

func (z *zeromq) receive(c chan bool) {
	for {
		bytes, err := z.socket.Recv(0)
		if err != nil {
			fmt.Println("recv:", err)
			c <- false
			return
		}
		for i := 0; i < len(z.hndChannels); i++ {
			idx := i
			go func() {
				z.hndChannels[idx] <- bytes
			}()
		}
		z.socket.Send([]byte("ACK"), 0)
	}
}

func (z *zeromq) Start(addr string) {
	go func() {
		for {
			err := z.listen(addr)
			if err == nil {
				c := make(chan bool)
				fmt.Println("ZeroMQ Receive started")
				go z.receive(c)
				<-c
			} else {
				time.Sleep(2500 * time.Millisecond)
			}
		}
	}()
}

func (z *zeromq) AddHandler(hnd Handler) {
	c := make(chan []byte)
	z.hndChannels = append(z.hndChannels, c)
	go func() {
		for {
			bytes := <-c
			hnd(bytes)
		}
	}()
}

func New() *zeromq {
	return new(zeromq)
}
