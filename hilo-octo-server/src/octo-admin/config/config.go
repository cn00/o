package config

import "github.com/BurntSushi/toml"

var config Config

type Config struct {
	Admin       AdminConfig       `toml:"admin"`
	Database    DatabaseConfig    `toml:"database"`
	OauthGoogle OauthGoogleConfig `toml:"oauth_google"`
	GCSProject  GCSProjectConfig  `toml:"gcs_project"`
}

type AdminConfig struct {
	CookieSecret string `toml:"cookie_secret"`
	Port         int    `toml:"port"`
	ReadOnly     bool   `toml:"read_only"`
	EnvCheck     bool   `toml:"env_check"`
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
}

type OauthGoogleConfig struct {
	ClientId     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectUrl  string `toml:"redirect_url"`
}

type GCSProjectConfig struct {
	ProjectId string `toml:"project_id"`
	Location  string `toml:"location"`
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
