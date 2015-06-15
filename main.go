package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/freahs/microhal"
	"github.com/mudler/sabbot/packages"
	"github.com/whyrusleeping/hellabot"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {

	botNick := "Sabbot"
	re, err := regexp.Compile(botNick + `\S`)
	channel := "#sabayon-dev"

	server := "irc.freenode.net:6667"
	brainFile := "Brain"
	learnEverything := 1
	messageOnJoin := 0

	var brain *microhal.Microhal
	if _, err := os.Stat(brainFile); os.IsNotExist(err) {
		brain = microhal.NewMicrohal(brainFile, 10)
	} else {
		brain = microhal.LoadMicrohal(brainFile)
	}

	brainIn, brainOut := brain.Start(5000*time.Millisecond, 250)

	nick := flag.String("nick", botNick, "nickname for the bot")
	serv := flag.String("server", server, "hostname and port for irc server to connect to")
	ichan := flag.String("chan", channel, "channel for bot to join")

	flag.Parse()

	irc, err := hbot.NewIrcConnection(*serv, *nick, false, false)
	if err != nil {
		panic(err)
	}
	var HALLearn = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			return mes.Command == "PRIVMSG" && mes.To == channel
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			//inputString := strings.Replace(mes.Content, botNick, "", 1)
			sanitizedInput := re.ReplaceAllLiteralString(mes.Content, "")
			brainIn <- sanitizedInput
			res := <-brainOut
			fmt.Printf("stupid response: %s\n", res)

			return false
		},
	}
	var HALAnswer = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			return strings.Contains(mes.Content, botNick) && mes.Command == "PRIVMSG" && mes.To == channel
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			//inputString := strings.Replace(mes.Content, botNick, "", 1)
			sanitizedInput := re.ReplaceAllLiteralString(mes.Content, "")
			brainIn <- sanitizedInput
			outputString := <-brainOut
			irc.Channels[mes.To].Say(mes.Name + ":" + outputString)
			return false
		},
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

			cmd := exec.Command("/usr/bin/eit", eitArgs)
			cmdReader, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
				os.Exit(1)
			}

			scanner := bufio.NewScanner(cmdReader)
			go func() {
				for scanner.Scan() {
					irc.Channels[m.To].Say(scanner.Text())
					time.Sleep(2000 * time.Millisecond)
				}
			}()

			err = cmd.Start()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
				//os.Exit(1)
			}

			err = cmd.Wait()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
				//os.Exit(1)
			}

			return true
		},
	}

	irc.AddTrigger(SearchPackage)
	irc.AddTrigger(SearchRevDeps)
	irc.AddTrigger(Eit)
	irc.AddTrigger(HALAnswer)
	if learnEverything == 1 {
		irc.AddTrigger(HALLearn)
	}
	// Start up bot
	irc.Start()

	// Join a channel
	mychannel := irc.Join(*ichan)
	if messageOnJoin == 1 {
		mychannel.Say("i'm here to serve")
	}
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
