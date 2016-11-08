package main

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	OAuth        OAuthConfig
	HTTP         HTTPConfig
	PollInterval int      `toml:"poll_interval"`
	QueryStats   []string `toml:"query_stats"`
	ConvertToF   bool     `toml:"convert_to_f"`
	OpenTSDB     OpenTSDBConfig
	Backend      string
	File         FileStateConfig
	Redis        RedisStateConfig
}

type HTTPConfig struct {
	Port int
}

// These are found at https://mytaglist.com/eth/oauth2_apps.html
type OAuthConfig struct {
	ID     string
	Secret string
}

type OpenTSDBConfig struct {
	Host          string
	Port          int
	MetricsPrefix string `toml:"metrics_prefix"`
}

type FileStateConfig struct {
	Filename string
}

type RedisStateConfig struct {
	Host string
	Port int
	Key  string
}

func ReadConfigFile(filename string) *Config {
	config := new(Config)
	tomlData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %s\n", err.Error())
	}
	_, err = toml.Decode(string(tomlData), config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s\n", err.Error())
	}
	return config
}
