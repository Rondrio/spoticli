package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"spotiCLI/spotify"
)

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Println(err)
		return
	}

	client := spotify.NewClient(cfg.ClientId, cfg.ClientSecret)

	client.Authorize(spotify.SCOPE_READ_PLAYBACK_POSITION)

}

type Config struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func loadConfig(path string) (*Config, error) {
	var config Config

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
