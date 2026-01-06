// Package config reads from and writes to config json in user's home dir
package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

func ResetConfig() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	template, err := os.ReadFile("./internal/config/template.json")
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, template, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func (c Config) SetUser(username string) error {
	c.CurrentUsername = username

	data, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func getConfigPath() (string, error) {
	const configFileName = "/.gatorconfig.json"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := homeDir + "/" + configFileName
	return configPath, nil
}

func Read() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
