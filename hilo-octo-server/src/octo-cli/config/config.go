package config

type Config struct {
	Api ApiConfig `toml:"api"`
	App AppConfig `toml:"app"`
	Oss OSSConfig `toml:"oss"`
}

type ApiConfig struct {
	BaseUrl string `toml:"base_url"`
}

type AppConfig struct {
	Id int `toml:"id"`
	Secret string `toml:"secret"`
}

type OSSConfig struct {
	AccessKey    string `toml:"AccessKey"`
	AccessSecret string `toml:"AccessSecret"`
	Endpoint     string `toml:"Endpoint"`
	Bucket       string `toml:"Bucket"`
	RootDir      string `toml:"RootDir"`
}
