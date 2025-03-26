package config


import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFile = "data/config.json"

// config struct for Ttracker data
type Config struct {
	Projects map[string]string `json:"project"`
	Active string `json:"active"`
}


// loadConfig reads and parses config.json
func LoadConfig() (Config, error) {
	var config Config

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return Config{Projects: map[string]string{}, Active: ""}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return Config{Projects: map[string]string{}, Active: ""}, err
	}

	json.Unmarshal(data, &config)
	return config, nil
}


// writes the updated config to file
func SaveConfig(config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error saving to Config.json:", err)
		return err
	}

	os.MkdirAll(filepath.Dir(configFile), os.ModePerm)
	os.WriteFile(configFile, data, 0644)
	return nil
}


