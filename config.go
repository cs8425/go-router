package main

import (
	"encoding/json"
	"os"
)

var config Config

// Config for server
type Config struct {
	Listen       string `json:"listen"`
	Log          string `json:"log"`
	Verb         int    `json:"verb"`
	UserDB       string `json:"userdb"`
//	Users        []User `json:"users"`
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path) // For read access.
	if err != nil {
		return err
	}
	defer file.Close()

//	config.Listen = ":8080"
//	config.Verb = 6

	config.Listen = *bind
	config.Verb = *verb
	config.UserDB = "user.json"

	return json.NewDecoder(file).Decode(config)
}

