package main

import (
	"errors"
	"fmt"
	irc "github.com/fluffle/goirc/client"
)

var ircConn *irc.Conn
var ircConnected bool

func StartIRC(server string, user string) {
	go func() {
		for {
			ircConn = irc.SimpleClient(user)
			//ircConn.SSL = true

			ircConn.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) {
				//conn.Join("#channel")
				ircConnected = true
			})

			quit := make(chan bool)
			ircConn.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) {
				ircConnected = false
				quit <- true
			})

			// Tell client to connect
			if err := ircConn.Connect(server); err != nil {
				fmt.Printf("Connection error: %s\n", err)
			}

			ircConn.Privmsg("willemvds", "STEVE!")

			// Wait for disconnect
			<-quit
		}
	}()
}

func IRCSendMessage(to string, msg string) error {
	if !ircConnected {
		return errors.New("Not connected to IRC Server")
	}
	ircConn.Privmsg(to, msg)
	return nil
}
