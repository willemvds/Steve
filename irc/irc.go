package irc

import (
	"errors"
	"fmt"
	goirc "github.com/fluffle/goirc/client"
)

type irc struct {
	conn      *goirc.Conn
	connected bool
}

func (i *irc) Start(server string, user string) {
	go func() {
		for {
			cfg := goirc.NewConfig(user)
			cfg.SSL = true
			cfg.Server = "irc.freenode.net:7000"
			cfg.NewNick = func(n string) string { return n + "^" }
			i.conn = goirc.Client(cfg)
			//i.conn = goirc.SimpleClient(user)
			//ircConn.SSL = true

			i.conn.HandleFunc("connected", func(conn *goirc.Conn, line *goirc.Line) {
				//conn.Join("#channel")
				i.connected = true
			})

			quit := make(chan bool)
			i.conn.HandleFunc("disconnected", func(conn *goirc.Conn, line *goirc.Line) {
				i.connected = false
				quit <- true
			})

			// Tell client to connect
			if err := i.conn.Connect(); err != nil {
				fmt.Printf("Connection error: %s\n", err)
			}

			i.SendMessage("willemvds", "STEVE!")

			// Wait for disconnect
			<-quit
		}
	}()
}

func (i *irc) SendMessage(to string, msg string) error {
	if !i.connected {
		return errors.New("Not connected to IRC Server")
	}
	i.conn.Privmsg(to, msg)
	return nil
}

func New() *irc {
	i := &irc{}
	return i
}
