package xmpp

import (
	"errors"
	"fmt"
	goxmpp "github.com/mattn/go-xmpp"
	"time"
)

type Handler func(*ChatView)

type xmpp struct {
	client      *goxmpp.Client
	connected   bool
	hndChannels []chan *ChatView
}

type ChatView struct {
	rem string
	typ string
	txt string
}

func (c *ChatView) GetRemote() string {
	return c.rem
}

func (c *ChatView) GetType() string {
	return c.typ
}

func (c *ChatView) GetText() string {
	return c.txt
}

func NewChatView(c goxmpp.Chat) *ChatView {
	return &ChatView{
		rem: c.Remote,
		typ: c.Type,
		txt: c.Text,
	}
}

func (x *xmpp) receive(c chan bool) {
	for {
		chat, err := x.client.Recv()
		if err != nil {
			x.connected = false
			c <- false
			return
		}
		switch v := chat.(type) {
		case goxmpp.Chat:
			cv := NewChatView(chat.(goxmpp.Chat))
			for i := 0; i < len(x.hndChannels); i++ {
				idx := i
				go func() {
					x.hndChannels[idx] <- cv
				}()
			}
		case goxmpp.Presence:
			fmt.Println(v.From, v.Show)
		}
	}
}

func (x *xmpp) Start(server string, user string, passwd string) {
	go func() {
		// keep running forever; reconnect if dc
		for {
			var err error
			x.client, err = goxmpp.NewClient(server, user, passwd, false)
			x.connected = (err == nil)
			if err == nil {
				c := make(chan bool)
				fmt.Println("Receive started")
				go x.receive(c)
				<-c
			} else {
				time.Sleep(2500 * time.Millisecond)
				fmt.Println(err)
			}
		}
	}()
}

func (x *xmpp) SendMessage(to string, msg string) error {
	if !x.connected {
		return errors.New("Not connected to XMPP Server")
	}
	x.client.Send(goxmpp.Chat{Remote: to, Type: "chat", Text: msg})
	return nil
}

func (x *xmpp) AddHandler(hnd Handler) {
	c := make(chan *ChatView)
	x.hndChannels = append(x.hndChannels, c)
	go func() {
		for {
			chatView := <-c
			hnd(chatView)
		}
	}()
}

func New() *xmpp {
	x := &xmpp{}
	return x
}
