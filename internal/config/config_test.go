package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	expected := tempHome + "/" + ".gatorconfig.json"
	actual, err := getConfigPath()
	if err != nil {
		t.Fatal(err)
	}

	if expected != actual {
		t.Errorf("expected %s but got %s", expected, actual)
	}
}

func TestRead(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	configFilePath, err := getConfigPath()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		DBURL:           "test",
		CurrentUsername: "alice",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(configFilePath, data, 0o664)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	if *cfg != *actual {
		t.Errorf("expected %v but got %v", cfg, actual)
	}
}

func TestReset(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := Config{
		DBURL:           "test",
		CurrentUsername: "alice",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(".gatorconfig.json", data, 0o664)
	if err != nil {
		t.Fatal(err)
	}

	templateConfig := getDefaultConfig()

	err = Reset()
	if err != nil {
		t.Fatal(err)
	}

	resetCfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	if templateConfig != *resetCfg {
		t.Errorf("expected %v but got %v", templateConfig, resetCfg)
	}
}

func TestSetUser(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := Config{
		DBURL: "test",
	}
	username := "alice"

	err := cfg.SetUser(username)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	if actual.CurrentUsername != username {
		t.Errorf("expected %s but got %s", username, actual.CurrentUsername)
	}
}
