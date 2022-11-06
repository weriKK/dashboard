package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type dashboardConfig struct {
	FeedConfig FeedConfig `yaml:"feeds"`
}

var cfg = &dashboardConfig{}

func LoadYAML(path string) error {

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config at %q: %w", path, err)
	}

	if err = yaml.Unmarshal(content, cfg); err != nil {
		return fmt.Errorf("not a valid YAML file: %w", err)
	}

	return nil
}
