package config

import "github.com/BurntSushi/toml"

var config Config

type Config struct {
	Api       ApiConfig       `toml:"api"`
	Database  DatabaseConfig  `toml:"database"`
	Metrics   MetricsConfig   `toml:"metrics"`
	CacheApps CacheAppsConfig `toml:"cache_apps"`
	CDN       CDNConfig       `toml:"cdn"`
}

type ApiConfig struct {
	Port              int     `toml:"port"`
	ReadOnly          bool    `toml:"read_only"`
	MinimumCliVersion float64 `toml:"minimum_cli_version"`
	EnvCheck          bool    `toml:"env_check"`
}

type DatabaseConfig struct {
	Master DatabaseServerConfig `toml:"master"`
	Slave  DatabaseServerConfig `toml:"slave"`
}

type DatabaseServerConfig struct {
	Addrs    string `toml:"addrs"`
	Dbname   string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	MaxOpen  int    `toml:"max_open"`
}

type MetricsConfig struct {
	Port int `toml:"port"`
}

type CacheAppsConfig struct {
	ListAPI []int `toml:"list_api"`
}

type CDNConfig struct {
	Default string            `toml:"default"`
	Apps    map[string]string `toml:"apps"`
}

func Init(file string) {
	_, err := toml.DecodeFile(file, &config)
	if err != nil {
		panic(err)
	}
}

func LoadConfig() *Config {
	return &config
}
