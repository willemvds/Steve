package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/willemvds/Steve/irc"
	"github.com/willemvds/Steve/math"
	"github.com/willemvds/Steve/xmpp"
)

var XMPPSendMessage func(string, string) error
var IRCSendMessage func(string, string) error

func Taxes(b []byte) {
	s := string(b)
	if strings.HasPrefix(s, "Taxes ") {
		parts := strings.SplitN(s, " ", 3)
		if len(parts) == 3 {
			target := parts[1]
			expr := parts[2]
			answer, err := math.Parse("Taxes", expr)
			if err == nil {
				XMPPSendMessage(target, fmt.Sprintf("Steve do taxes for %s: %s = %d moneys to taxman", target, expr, answer))
			}
		}
	}
}

func Print(cv *xmpp.ChatView) {
	fmt.Println("Received from", cv.GetRemote(), ":", cv.GetText())
}

func Log(cv *xmpp.ChatView) {
	log.Print("Received from", cv.GetRemote(), ":", cv.GetText())
}

func Reply(cv *xmpp.ChatView) {
	if len(cv.GetText()) > 0 {
		XMPPSendMessage(cv.GetRemote(), fmt.Sprintf("ACK: %s", cv.GetText()))
	}
}

func ForwardToIRC(cv *xmpp.ChatView) {
	tokens := strings.SplitN(cv.GetText(), " ", 2)
	if len(tokens) == 2 {
		IRCSendMessage(strings.TrimSpace(tokens[0]), tokens[1])
	}
}

func UName(cv *xmpp.ChatView) {
	if cv.GetText() == "uname" {
		cmd := exec.Command("uname", "-a")
		output, err := cmd.CombinedOutput()
		if err != nil {
			XMPPSendMessage(cv.GetRemote(), fmt.Sprintf("%s", err))
		} else {
			XMPPSendMessage(cv.GetRemote(), string(output))
		}
	}
}

func DoMath(cv *xmpp.ChatView) {
	if strings.HasSuffix(cv.GetText(), "= ?") {
		expr := strings.TrimSpace(cv.GetText()[0 : len(cv.GetText())-3])
		answer, err := math.Parse("STEVE!", expr)
		if err != nil {
			XMPPSendMessage(cv.GetRemote(), "Steve not know!")
			return
		}
		XMPPSendMessage(cv.GetRemote(), fmt.Sprintf("%d", answer))
	}
}

func main() {
	user := flag.String("user", "", "gtalk username")
	passwd := flag.String("passwd", "", "gtalk password")
	flag.Parse()

	if len(*user) == 0 || len(*passwd) == 0 {
		flag.PrintDefaults()
		return
	}

	freenode := irc.New()
	freenode.Start("irc.freenode.org", "monkeysteve")
	IRCSendMessage = func(target string, message string) error {
		return freenode.SendMessage(target, message)
	}

	gtalk := xmpp.New()
	gtalk.Start("talk.google.com:443", *user, *passwd)
	XMPPSendMessage = func(target string, message string) error {
		return gtalk.SendMessage(target, message)
	}
	gtalk.AddHandler(Print)
	gtalk.AddHandler(Log)
	gtalk.AddHandler(Reply)
	gtalk.AddHandler(ForwardToIRC)
	gtalk.AddHandler(UName)
	gtalk.AddHandler(DoMath)

	for {
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')
		if err != nil {
			continue
		}
		line = strings.TrimRight(line, "\n")
		tokens := strings.SplitN(line, " ", 2)
		if len(tokens) == 2 {
			err := XMPPSendMessage(tokens[0], tokens[1])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
