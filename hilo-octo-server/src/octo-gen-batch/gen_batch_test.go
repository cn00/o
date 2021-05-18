package main

import (
	"flag"
	"testing"

	"octo-api/config"
	"octo-api/service"
	"octo/models"
	"octo/utils"

	"github.com/stretchr/testify/assert"
)

var confFile string
var appId int
var isGenNull bool
var isLocal bool = false

func DBSetup() {
	flag.StringVar(&confFile, "conf", "config.tml", "config toml file")
	flag.Parse()
	utils.RandSeed()

	config.Init(confFile)
	conf := config.LoadConfig()

	service.Setup(service.Config{
		CacheAppsListAPI: conf.CacheApps.ListAPI,
	})

	if err := models.Setup(models.Config{
		ReadOnly:               false,
		DatabaseMasterAddrs:    conf.Database.Master.Addrs,
		DatabaseMasterDbname:   conf.Database.Master.Dbname,
		DatabaseMasterUser:     conf.Database.Master.User,
		DatabaseMasterPassword: conf.Database.Master.Password,
		DatabaseSlaveAddrs:     conf.Database.Slave.Addrs,
		DatabaseSlaveDbname:    conf.Database.Slave.Dbname,
		DatabaseSlaveUser:      conf.Database.Slave.User,
		DatabaseSlavePassword:  conf.Database.Slave.Password,
	}); err != nil {
		panic(err)
	}
	appDao = &models.AppDao{}
	fileDao = models.NewFileDao()
	fileUrlDao = &models.FileUrlDao{}
	resourceDao = models.NewResourceDao()
	resourceUrlDao = &models.ResourceUrlDao{}
}

func TestUpdateGenerationFromUrlsWithApp(t *testing.T) {
	if isLocal {
		DBSetup()

		err := UpdateGenerationFromUrlsWithApp(0)
		assert.Equal(t, nil, err, "All App UpdateGenerationFromUrlsWithApp")

		err = UpdateGenerationFromUrlsWithApp(1)
		assert.Equal(t, nil, err, "App UpdateGenerationFromUrlsWithApp")
	}
}

func TestUpdateGenerationOfGenerationNullData(t *testing.T) {
	if isLocal {
		DBSetup()

		err := UpdateGenerationAndUploadVersionIdOfNullData(0)
		assert.Equal(t, nil, err, "All App UpdateGenerationOfGenerationNullData")

		err = UpdateGenerationAndUploadVersionIdOfNullData(1)
		assert.Equal(t, nil, err, "App UpdateGenerationOfGenerationNullData")
	}
}
