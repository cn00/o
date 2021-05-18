package config

type Config struct {
	Api ApiConfig `toml:"api"`
	App AppConfig `toml:"app"`
}

type ApiConfig struct {
	BaseUrl string `toml:"base_url"`
}

type AppConfig struct {
	Id int `toml:"id"`
}
