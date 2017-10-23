package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	FrontendBootstrapPage string `json:"frontend_bootstrap_page"`
	SimulatedGPIOEnabled  bool   `json:"simulated_gpio_enabled"`
}

func Load(configFile string) Config {
	// todo: validate required config
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("couldn't file config file: %v", configFile)
	}
	var config Config
	json.NewDecoder(file).Decode(&config)
	return config
}
