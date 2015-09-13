package main

import (
	"github.com/freahs/microhal"
	"github.com/mudler/sabbot/packages"
	"github.com/whyrusleeping/hellabot"

	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const cmdPrefix = "!"

func AddTriggers(irc *hbot.IrcCon, config hbot.Config) {

	/* Start microhal */
	//microhal too unstable
	HAL := true
	brainFile := "Brain"
	learnEverything := true
	markovOrder := 4

	if HAL {
		re, _ := regexp.Compile(config.Nick + `\S`)

		var brain *microhal.Microhal

		if _, err := os.Stat(brainFile); os.IsNotExist(err) {
			brain = microhal.NewMicrohal(brainFile, markovOrder)
		} else {
			brain = microhal.LoadMicrohal(brainFile)
		}
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
				if len(sanitizedInput) > markovOrder {
					brainIn <- sanitizedInput
					res := <-brainOut
					fmt.Printf("stupid response: %s\n", res)
				}

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
				if len(sanitizedInput) > markovOrder {

					brainIn <- sanitizedInput
					outputString := <-brainOut
					irc.Channels[m.To].Say(m.Name + ":" + outputString)
				}
				return false
			},
		}
		irc.AddTrigger(HALAnswer)
		if learnEverything {
			irc.AddTrigger(HALLearn)
		}
	}
	/* End microhal */

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
				irc.Channels[m.To].Say("Please, be more specific next time")
				return
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
			if strings.Contains(m.Content, cmdPrefix+"search") {
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
			if strings.Contains(m.Content, cmdPrefix+"rdeps") {
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
			if m.Content == cmdPrefix+"latest" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			go Search(irc, m, "SearchPackage")
			return false
		},
	}

	var ShowURLTitles = &hbot.Trigger{
		func(m *hbot.Message) bool {
			return strings.Contains(m.Content, "http://") || strings.Contains(m.Content, "https://") || strings.Contains(m.Content, "www.")
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			title := UrlTitle(m.Content)
			irc.Channels[m.To].Say(title)
			return false
		},
	}

	var DDG = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if strings.Contains(m.Content, cmdPrefix+"ddg") {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			ddgArgs := strings.Replace(m.Content, cmdPrefix+"ddg", "", 1)
			go irc.Channels[m.To].Say(SearchCmd(ddgArgs))
			return true
		},
	}

	var access = []string{"joost_op", "mudler", "Enlik"}
	var Eit = &hbot.Trigger{
		func(m *hbot.Message) bool {
			if strings.Contains(m.Content, cmdPrefix+"eit") {
				for _, s := range access {
					if m.From == s {
						return true
					}
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {

			eitArgs := strings.Replace(m.Content, cmdPrefix+"eit", "", 1)

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
			if m.Content == cmdPrefix+"help" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			irc.Channels[m.To].Say(cmdPrefix + "search package - searches a package in https://packages.sabayon.org/")
			irc.Channels[m.To].Say(cmdPrefix + "rdeps package - searches a package reverse deps in https://packages.sabayon.org/")
			irc.Channels[m.To].Say(cmdPrefix + "latest - show the latest compiled packages in https://packages.sabayon.org/")
			irc.Channels[m.To].Say(cmdPrefix + "eit args - Calls eit with given args and print the output")
			irc.Channels[m.To].Say(cmdPrefix + "ddg args - Execute a search via DuckDuckGo, bangs included (e.g. " + cmdPrefix + "ddg !google kittens)")

			irc.Channels[m.To].Say("...try to write a VALID url :)")

			return true
		},
	}

	var PrintMessages = &hbot.Trigger{
		func(m *hbot.Message) bool {
			return true
		},
		func(irc *hbot.IrcCon, m *hbot.Message) bool {
			fmt.Println(m.To + ": " + m.Content)
			return true
		},
	}

	irc.AddTrigger(SearchPackage)
	irc.AddTrigger(SearchRevDeps)
	irc.AddTrigger(LatestPackage)
	irc.AddTrigger(DDG)
	irc.AddTrigger(ShowURLTitles)
	irc.AddTrigger(Eit)

	irc.AddTrigger(Help)
	irc.AddTrigger(PrintMessages)

}
