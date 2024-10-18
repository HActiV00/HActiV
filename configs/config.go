package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	FilterPatterns []string `json:"filter_patterns"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}
