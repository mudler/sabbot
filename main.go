package main

import (
	"github.com/mudler/sabbot/packages"
	"flag"
	"fmt"
	"github.com/whyrusleeping/hellabot"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	channel := "#spike-pentesting-dev"
	bot_nick := "Sabbot"
	server := "irc.freenode.net:6667"

	nick := flag.String("nick", bot_nick, "nickname for the bot")
	serv := flag.String("server", server, "hostname and port for irc server to connect to")
	ichan := flag.String("chan", channel, "channel for bot to join")

	flag.Parse()

	irc, err := hbot.NewIrcConnection(*serv, *nick, false, false)
	if err != nil {
		panic(err)
	}

	var Search = func(irc *hbot.IrcCon, mes *hbot.Message, s string) {
		max := 3
		var search []packages.Package
		var query string
		words := strings.Fields(mes.Content)
		irc.Channels[mes.To].Say("Searching, be patient boy")
		if s == "SearchPackage" {
			search, query = packages.Search(words[1])
		} else if s == "SearchRevDeps" {
			search, query = packages.ReverseDeps(words[1])
		}
		irc.Channels[mes.To].Say("Showing results for " + query + " limited to " + strconv.Itoa(max) + " results")
		if len(search) < max {
			max = len(search)
		}
		for i := 0; i < max; i++ {
			irc.Channels[mes.To].Say(search[i].String())
			time.Sleep(1000 * time.Millisecond)
		}
	}

	var SearchPackage = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if strings.Contains(mes.Content, "-search") {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			Search(irc, mes, "SearchPackage")
			return false
		},
	}

	var SearchRevDeps = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if strings.Contains(mes.Content, "-rdeps") {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			Search(irc, mes, "SearchRevDeps")
			return false
		},
	}

	var access = []string{"joost_op", "mudler", "Enlik"}
	var Eit = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if strings.Contains(m.Content, "-eit") {
				for _, s := range access {
					if m.From == s {
						return true
					}
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {

			eitArgs := strings.Replace(m.Content, "-eit", "", 1)
			out, err := exec.Command("/usr/bin/eit", eitArgs).Output()
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}
			output := string(out[:])
			irc.Channels[m.To].Say(output)

			return true
		},
	}

	irc.AddTrigger(SearchPackage)
	irc.AddTrigger(SearchRevDeps)
	irc.AddTrigger(Eit)

	// Start up bot
	irc.Start()

	// Join a channel
	mychannel := irc.Join(*ichan)
	mychannel.Say("i'm here to serve")

	// Read off messages from the server
	for mes := range irc.Incoming {
		if mes == nil {
			fmt.Println("Disconnected.")
			return
		}
		// Log raw message struct
		fmt.Println(mes)
	}
	fmt.Println("Bot shutting down.")
}
