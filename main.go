package main

import (
	"bufio"
	"fmt"
	"os"
	// "internal/pokecache"
)

func main() {
	var conf config
	conf = config{
		Next:     "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Previous: "",
	}

	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	for {
		var userInput string
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			userInput = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}

		words := cleanInput(userInput)
		command := words[0]
		args := words[1:]

		if err := commands[command].callback(&conf, args); err != nil {
			fmt.Printf("error: %s\n", err.Error())
		}
	}
}
