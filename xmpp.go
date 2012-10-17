package main

import (
	"errors"
	"fmt"
	"github.com/mattn/go-xmpp"
	"time"
)

var xmppClient *xmpp.Client
var xmppConnected bool
var xmppChannels []chan *ChatView

type XMPPHandler func(*ChatView)

func init() {
	xmppChannels = make([]chan *ChatView, 0)
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

func NewChatView(c *xmpp.Chat) *ChatView {
	return &ChatView{
		rem: c.Remote,
		typ: c.Type,
		txt: c.Text,
	}
}

func xmppReceive(c chan bool) {
	for {
		chat, err := xmppClient.Recv()
		if err != nil {
			xmppConnected = false
			c <- false
			return
		}
		cv := NewChatView(&chat)
		for i := 0; i < len(xmppChannels); i++ {
			idx := i
			go func() {
				xmppChannels[idx] <- cv
			}()
		}
	}
}

func StartXMPP(server string, user string, passwd string) {
	go func() {
		// keep running forever; reconnect if dc
		for {
			var err error
			xmppClient, err = xmpp.NewClient(server, user, passwd)
			xmppConnected = (err == nil)
			if err == nil {
				c := make(chan bool)
				fmt.Println("Receive started")
				go xmppReceive(c)
				<-c
			} else {
				time.Sleep(1 * time.Second)
				fmt.Println(err)
			}
		}
	}()
}

func XMPPSendMessage(to string, msg string) error {
	if !xmppConnected {
		return errors.New("Not connected to XMPP Server")
	}
	xmppClient.Send(xmpp.Chat{Remote: to, Type: "chat", Text: msg})
	return nil
}

func NewXMPPReceiver() chan *ChatView {
	c := make(chan *ChatView)
	xmppChannels = append(xmppChannels, c)
	return c
}

func AddXMPPHandler(hnd XMPPHandler) {
	c := make(chan *ChatView)
	xmppChannels = append(xmppChannels, c)
	go func() {
		for {
			chatView := <-c
			hnd(chatView)
		}
	}()
}
