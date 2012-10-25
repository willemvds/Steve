package main

import (
	"./math"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Print(cv *ChatView) {
	fmt.Println("Received from", cv.GetRemote(), ":", cv.GetText())
}

func Log(cv *ChatView) {
	log.Print("Received from", cv.GetRemote(), ":", cv.GetText())
}

func Reply(cv *ChatView) {
	XMPPSendMessage(cv.GetRemote(), fmt.Sprintf("ACK: %s", cv.GetText()))
}

func ForwardToIRC(cv *ChatView) {
	tokens := strings.SplitN(cv.GetText(), " ", 2)
	if len(tokens) == 2 {
		IRCSendMessage(strings.TrimSpace(tokens[0]), tokens[1])
	}
}

func UName(cv *ChatView) {
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

func DoMath(cv *ChatView) {
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

	go StartIRC("irc.freenode.org", "monkeysteve")

	if len(*user) == 0 || len(*passwd) == 0 {
		flag.PrintDefaults()
		return
	}

	StartXMPP("talk.google.com:443", *user, *passwd)

	AddXMPPHandler(Print)
	AddXMPPHandler(Log)
	AddXMPPHandler(Reply)
	AddXMPPHandler(ForwardToIRC)
	AddXMPPHandler(UName)
	AddXMPPHandler(DoMath)

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

	select {}
}
