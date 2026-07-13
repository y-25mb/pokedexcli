package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/y-25mb/pokedexcli/internal"
)

type Pokedex struct {
	caught map[string]Pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	Next     string
	Previous string
}

var cache = internal.NewCache(100 * time.Second)
var pokedex = Pokedex{
	caught: map[string]Pokemon{},
}

func getCommands() map[string]cliCommand {
	var commands = map[string]cliCommand{
		"help": cliCommand{
			name:        "help",
			description: "Displays help message",
			callback:    commandHelp,
		},
		"map": cliCommand{
			name:        "map",
			description: "Displays a list of areas. Navigate forward in the list.",
			callback:    commandMap,
		},
		"mapb": cliCommand{
			name:        "mapb",
			description: "Displays a list of areas. Navigates backward in the list.",
			callback:    commandMapb,
		},
		"explore": cliCommand{
			name:        "explore",
			description: "Displays information about an area.",
			callback:    commandExplore,
		},
		"catch": cliCommand{
			name:        "catch",
			description: "Attempts to catch a pokemon.",
			callback:    commandCatch,
		},
		"pokedex": cliCommand{
			name:        "pokedex",
			description: "List caught pokemon in pokedex.",
			callback:    commandListPokedex,
		},
		"inspect": cliCommand{
			name:        "inspect",
			description: "Inspect caught pokemon.",
			callback:    commandInspect,
		},
		"exit": cliCommand{
			name:        "exit",
			description: "exits program",
			callback:    commandExit,
		},
	}

	return commands
}

func cleanInput(text string) []string {
	out := []string{}

	text = strings.ToLower(text)
	text = strings.TrimSpace(text)

	out = strings.Fields(text)

	return out
}

func commandExit(conf *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

func commandHelp(conf *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	for key, value := range getCommands() {
		fmt.Println(key + ": " + value.description)
	}
	return nil
}

func commandMap(conf *config, args []string) error {
	if conf.Previous == "" {
		fmt.Println("you're on the first page")
	} else if conf.Next == "" {
		fmt.Println("you're on the last page")
		return nil
	}

	data, err := getMapData(conf.Next)
	if err != nil {
		return err
	}

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	conf.Next = data.Next
	conf.Previous = data.Previous

	return nil
}

func commandMapb(conf *config, args []string) error {
	if conf.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	} else if conf.Next == "" {
		fmt.Println("you're on the last page")
	}

	data, err := getMapData(conf.Previous)
	if err != nil {
		return err
	}

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}

	conf.Next = data.Next
	conf.Previous = data.Previous

	return nil
}

func commandCatch(conf *config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Usage: catch <area>")
	}

	pokemonToCatch := args[0]
	pokemon, err := getPokemon("https://pokeapi.co/api/v2/pokemon/" + pokemonToCatch)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	randomInt := rand.IntN(pokemon.BaseExperience)
	if randomInt >= 0 && randomInt < 30 {
		fmt.Printf("%s was caught!\n", pokemon.Name)
		pokedex.caught[pokemon.Name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}

	return nil
}

func commandListPokedex(conf *config, args []string) error {
	if len(pokedex.caught) == 0 {
		fmt.Println("No pokemon in pokedex.")
		return nil
	}
	fmt.Println("Pokemon caught:")
	for key, _ := range pokedex.caught {
		fmt.Printf(" - %s\n", key)
	}
	return nil
}

func commandInspect(conf *config, args []string) error {
	pokemonName := args[0]
	pokemonStruct, ok := pokedex.caught[pokemonName]
	if !ok {
		fmt.Println("Pokemon not caught!")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemonStruct.Name)
	fmt.Printf("Height: %v\n", pokemonStruct.Height)
	fmt.Printf("Weight: %v\n", pokemonStruct.Weight)

	fmt.Printf("Stats:\n")
	for _, stat := range pokemonStruct.Stats {
		fmt.Printf("  -%v: %v\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Printf("Name: %s\n", pokemonStruct.Name)

	for _, pokemonType := range pokemonStruct.Types {
		fmt.Printf("  - %v\n", pokemonType.Type.Name)
	}

	return nil
}

func commandExplore(conf *config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("Usage: explore <area>")
	}

	area := args[0]
	// if !ok || area == "" {
	// 	return fmt.Errorf("invalid location area parameter")
	// }

	encounters, err := getEncounters("https://pokeapi.co/api/v2/location-area/" + area)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\nFound Pokemon:\n", area)
	for _, encounter := range encounters {
		fmt.Println(" -", encounter.Pokemon.Name)
	}

	return nil
}

func getMapData(url string) (areas, error) {
	var zero areas
	res, err := getAndUnmarshal[areas](url)
	if err != nil {
		return zero, err
	}

	return res, nil
}

func getPokemon(url string) (Pokemon, error) {
	var zero Pokemon
	res, err := getAndUnmarshal[Pokemon](url)
	if err != nil {
		return zero, err
	}

	return res, nil
}

func getEncounters(url string) ([]pokemonEncounter, error) {
	var zero []pokemonEncounter
	res, err := getAndUnmarshal[locationArea](url)
	if err != nil {
		return zero, err
	}

	return res.PokemonEncounters, nil
}

func getAndUnmarshal[T any](url string) (T, error) {
	var zero T
	var dataStruct T

	cacheEntry, ok := cache.Get(url)
	if ok {
		// fmt.Println("Cache hit!")
		err := json.Unmarshal(cacheEntry, &dataStruct)
		if err != nil {
			return zero, fmt.Errorf("(Cache) Issue decoding area json: %w", err)
		}
		return dataStruct, nil
	}
	// fmt.Println("Cache miss!")

	// fmt.Println("GETting", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return zero, fmt.Errorf("Issue creating http request for location: %w", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return zero, fmt.Errorf("Issue performing GET request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return zero, fmt.Errorf("(Cache) Issue reading res body: %w", err)
	}

	cache.Add(url, data)

	err = json.Unmarshal(data, &dataStruct)
	if err != nil {
		return zero, fmt.Errorf("Issue decoding json: %w", err)
	}

	return dataStruct, nil
}
