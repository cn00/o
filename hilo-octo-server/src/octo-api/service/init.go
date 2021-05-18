package service

var conf Config

type Config struct {
	CacheAppsListAPI []int
}

func Setup(c Config) {
	conf = c
}

func listAPICacheEnabled(targetAppId int) bool {
	for _, appId := range conf.CacheAppsListAPI {
		if appId == targetAppId {
			return true
		}
	}
	return false
}
