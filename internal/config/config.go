// Package config reads from and writes to config json in user's home dir
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

func getDefaultConfig() Config {
	return Config{
		DBURL: "postgres://postgres:@localhost:5432/gator?sslmode=disable",
	}
}

func getConfigPath() (string, error) {
	const configFileName = ".gatorconfig.json"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home dir: %v", err)
	}

	configPath := filepath.Join(homeDir, configFileName)
	return configPath, nil
}

func Reset() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("could not get config file path: %v", err)
	}

	templateConfig := getDefaultConfig()
	templateJSON, err := json.Marshal(templateConfig)
	if err != nil {
		return fmt.Errorf("could not marshal template: %v", err)
	}

	err = os.WriteFile(configPath, templateJSON, 0o644)
	if err != nil {
		return fmt.Errorf("could not write template json to config file: %v", err)
	}

	return nil
}

func (c *Config) write() error {
	data, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return fmt.Errorf("could not marshal config struct: %v", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("could not get config file path: %v", err)
	}

	err = os.WriteFile(configPath, data, 0o644)
	if err != nil {
		return fmt.Errorf("could not write template json to config file: %v", err)
	}

	return nil
}

func (c *Config) SetDB(url string) error {
	c.DBURL = url

	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUsername = username

	err := c.write()
	if err != nil {
		return err
	}

	return nil
}

func Read() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not get config file path: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %v", err)
	}

	return &cfg, nil
}
