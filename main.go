package main

import (
	"bufio"
	"fmt"
	"github.com/freahs/microhal"
	"github.com/mudler/hellabot"
	"github.com/mudler/sabbot/packages"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	configFile := "./config.json"
	HAL := false
	brainFile := "Brain"
	learnEverything := true
	var brain *microhal.Microhal

	if _, err := os.Stat(brainFile); os.IsNotExist(err) {
		brain = microhal.NewMicrohal(brainFile, 6)
	} else {
		brain = microhal.LoadMicrohal(brainFile)
	}

	irc, config, err := hbot.NewIrcConnectionFromJSON(configFile)
	fmt.Println("Loading from " + configFile)
	if err != nil {
		panic(err)
	}
	re, reerr := regexp.Compile(config.Nick + `\S`)
	if reerr != nil {
		panic(err)
	}

	//microhal too unstable
	if HAL {
		brainIn, brainOut := brain.Start(10000*time.Millisecond, 250)

		var HALLearn = &hbot.Trigger{
			func(m *hbot.Message) bool {
				if m.Command == "PRIVMSG" {
					return true
				}
				return false
			},
			func(irc *hbot.IrcCon, m *hbot.Message) bool {
				//inputString := strings.Replace(mes.Content, botNick, "", 1)
				sanitizedInput := re.ReplaceAllLiteralString(m.Content, "")
				brainIn <- sanitizedInput
				res := <-brainOut
				fmt.Printf("stupid response: %s\n", res)

				return false
			},
		}
		var HALAnswer = &hbot.Trigger{
			func(m *hbot.Message) bool {
				return strings.Contains(m.Content, config.Nick) && m.Command == "PRIVMSG"
			},
			func(irc *hbot.IrcCon, m *hbot.Message) bool {
				//inputString := strings.Replace(mes.Content, botNick, "", 1)
				sanitizedInput := re.ReplaceAllLiteralString(m.Content, "")
				brainIn <- sanitizedInput
				outputString := <-brainOut
				irc.Channels[m.To].Say(m.Name + ":" + outputString)
				return false
			},
		}
		irc.AddTrigger(HALAnswer)
		if learnEverything {
			irc.AddTrigger(HALLearn)
		}
	}

	var Search = func(irc *hbot.IrcCon, m *hbot.Message, s string) {
		max := 3
		var search []packages.Package
		var query string
		words := strings.Fields(m.Content)
		irc.Channels[m.To].Say("Searching, be patient boy")
		if s == "SearchPackage" {
			if len(words) == 2 {
				search, query = packages.Search(words[1])
			} else {
				search, query = packages.Search("")
			}
		} else if s == "SearchRevDeps" {
			if len(words) == 2 {
				search, query = packages.ReverseDeps(words[1])
			} else {
				search, query = packages.ReverseDeps("")
			}
		}
		irc.Channels[m.To].Say("Showing results for " + query + " limited to " + strconv.Itoa(max) + " results")
		if len(search) < max {
			max = len(search)
		}
		for i := 0; i < max; i++ {
			irc.Channels[m.To].Say(search[i].String())
			time.Sleep(1000 * time.Millisecond)
		}
	}

	var SearchPackage = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if strings.Contains(m.Content, "-search") {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			Search(irc, m, "SearchPackage")
			return false
		},
	}

	var SearchRevDeps = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if strings.Contains(m.Content, "-rdeps") {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			go Search(irc, m, "SearchRevDeps")
			return false
		},
	}

	var LatestPackage = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if m.Content == "-latest" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			go Search(irc, m, "SearchPackage")
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

	var Help = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if m.Content == "-help" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			irc.Channels[m.To].Say("-search package - searches a package in https://packages.sabayon.org/")
			irc.Channels[m.To].Say("-rdeps package - searches a package reverse deps in https://packages.sabayon.org/")
			irc.Channels[m.To].Say("-latest - show the latest compiled packages in https://packages.sabayon.org/")
			irc.Channels[m.To].Say("-eit args - Calls eit with given args and print the output")
			return true
		},
	}

	irc.AddTrigger(SearchPackage)
	irc.AddTrigger(SearchRevDeps)
	irc.AddTrigger(LatestPackage)
	irc.AddTrigger(Help)
	irc.AddTrigger(Eit)

	// Start up bot
	irc.Start()

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
