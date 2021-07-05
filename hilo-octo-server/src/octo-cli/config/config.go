package config

type Config struct {
	Api ApiConfig `toml:"api"`
	App AppConfig `toml:"app"`
	Gcs GcsConfig `toml:"gcs"`
	Oss OSSConfig `toml:"oss"`
	Cos CosConfig `toml:"cos"`
}

type ApiConfig struct {
	BaseUrl string `toml:"base_url"`
}

type AppConfig struct {
	Id     int    `toml:"id"`
	Secret string `toml:"secret"`
}

type GcsConfig struct {
	ProjectID  string `toml:"projectID"`
	BucketName string `toml:"bucketName"`
	Location   string `toml:"location"`
}

type OSSConfig struct {
	AccessKey    string `toml:"AccessKey"`
	AccessSecret string `toml:"AccessSecret"`
	Endpoint     string `toml:"Endpoint"`
	Bucket       string `toml:"Bucket"`
	RootDir      string `toml:"RootDir"`
}

type CosConfig struct {
	BaseUrl   string `toml:"BaseUrl"`
	SecretID  string `toml:"SecretID"`
	SecretKey string `toml:"SecretKey"`
	RootDir   string `toml:"RootDir"`
}
