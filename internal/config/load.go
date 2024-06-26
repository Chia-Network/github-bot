package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads config from the given path
func LoadConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	config.SkipUsersMap = map[string]bool{}
	for _, user := range config.SkipUsers {
		config.SkipUsersMap[user] = true
	}

	return config, nil
}
