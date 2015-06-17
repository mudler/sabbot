package main

import (
	"github.com/mudler/hellabot"

	"fmt"
)

func main() {
	configFile := "./config.json"

	config, _ := hbot.LoadConfig(configFile)

	irc, config, err := hbot.NewIrcConnectionFromJSON(config)
	fmt.Println("Loading from " + configFile)
	if err != nil {
		panic(err)
	}

	AddTriggers(irc, config)

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
