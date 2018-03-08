package main

import (
	"encoding/json"
	"log"
	"os"
)

// todo: validate required config

type Config struct {
	Port             uint16    `json:"port"`
	FrontendRoot     string    `json:"frontend_root"`
	MediaDirectories []string  `json:"media_directories"`
	Controls         []Control `json:"controls"`
}

type Control struct {
	Control string  `json:"control"`
	Type    string  `json:"type"`
	Pins    []uint8 `json:"pins"`
}

func LoadConfig(configFile string) Config {
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("couldn't file config file: %v", configFile)
	}
	var config Config
	json.NewDecoder(file).Decode(&config)
	return config
}

func FlattenedPins(controls []Control) []uint8 {
	pins := make([]uint8, 0, len(controls))
	for _, control := range controls {
		for _, pin := range control.Pins {
			pins = append(pins, pin)
		}
	}
	return pins
}
