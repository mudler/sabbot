package main

import (
	"github.com/whyrusleeping/hellabot"

	"fmt"
)

func main() {
	var configFile string
	var irc *hbot.IrcCon
	var config hbot.Config
	var err error

	configFile = "./config.json"

	config, _ = hbot.LoadConfig(configFile)

	irc, config, err = hbot.NewIrcConnectionFromJSON(config)
	fmt.Println("Loading from " + configFile)
	if err != nil {
		panic(err)
	}

	AddTriggers(irc, config)

	// Start up bot
	irc.Start()
	for true {
	}

}
