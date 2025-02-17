package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OpenCommand []string `yaml:"open_command"`
	ViewMode    int      `yaml:"view_mode"`
	MaxDepth    int      `yaml:"max_depth"`
	CacheLimit  int      `yaml:"cache_limit"`
	WGMax       int      `yaml:"wg_max"`
}

var configPath = []string{
	"./parax2_config.yaml",
	filepath.Join(os.Getenv("HOME"), ".parax2_config.yaml"),
}

func loadConfig() *Config {
	for _, path := range configPath {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		config := Config{
			OpenCommand: nil,
			ViewMode:    -255,
			MaxDepth:    -255,
			CacheLimit:  -255,
			WGMax:       -255,
		}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			continue
		}

		return &config
	}

	return nil
}
