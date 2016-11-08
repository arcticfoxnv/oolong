package main

import (
	"testing"
)

func TestConfigFileGlobal(t *testing.T) {
	config := ReadConfigFile("oolong.toml.example")
	if config.PollInterval == 0 {
		t.Fail()
	}

	if len(config.QueryStats) == 0 {
		t.Fail()
	}

	if config.Backend == "" {
		t.Fail()
	}
}

func TestConfigFileOAuth(t *testing.T) {
	config := ReadConfigFile("oolong.toml.example")
	if config.OAuth.ID == "" {
		t.Fail()
	}

	if config.OAuth.Secret == "" {
		t.Fail()
	}
}

func TestConfigFileOpenTSDB(t *testing.T) {
	config := ReadConfigFile("oolong.toml.example")
	if config.OpenTSDB.Host == "" {
		t.Fail()
	}

	if config.OpenTSDB.Port == 0 {
		t.Fail()
	}

	if config.OpenTSDB.MetricsPrefix == "" {
		t.Fail()
	}
}

func TestConfigFileFile(t *testing.T) {
	config := ReadConfigFile("oolong.toml.example")
	if config.File.Filename == "" {
		t.Fail()
	}
}

func TestConfigFileRedis(t *testing.T) {
	config := ReadConfigFile("oolong.toml.example")
	if config.Redis.Host == "" {
		t.Fail()
	}

	if config.Redis.Port == 0 {
		t.Fail()
	}

	if config.Redis.Key == "" {
		t.Fail()
	}
}
