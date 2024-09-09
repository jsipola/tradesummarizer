package helpers

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Name   int `json:"name"`
	Isin   int `json:"isin"`
	Type   int `json:"type"`
	Ticker int `json:"ticker"`
	Date   int `json:"date"`
	Shares int `json:"shares"`
	Amount int `json:"amount"`
}

func ReadJsonConfig(path string) Config {
	config := createDefaultConfig()

	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading config file")
		writeConfigToFile(config)
		return config
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		fmt.Println("Error parsing config file")
		return config
	}

	return config
}

func createDefaultConfig() Config {
	return Config{
		Name:   2,
		Isin:   3,
		Type:   1,
		Ticker: 4,
		Date:   6,
		Shares: 8,
		Amount: 22,
	}
}

func writeConfigToFile(cfg Config) {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling config")
		return
	}
	file, err := os.Create("default.json")
	if err != nil {
		fmt.Println("Error creating default config file")
		return
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error saving config to file")
	}
}
