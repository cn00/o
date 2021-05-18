package main

import (
	"flag"
	"log"
	"octo-api/config"
	"octo/models"
	"octo/utils"
	"time"
)

var AppDao = &models.AppDao{}
var VersionDao = &models.VersionDao{}
var FileDao = models.NewFileDao()
var FileUrlDao = &models.FileUrlDao{}
var ResourceDao = models.NewResourceDao()
var ResourceUrlDao = &models.ResourceUrlDao{}

// 30 days
var DeleteTerm = 30 * 24 * 60

// Hard Delete for Soft Delete that files, file_urls, resources, resource_urls, version
func main() {
	log.Println("Version Hard Delete Batch Start")
	start := time.Now()
	var confFile string
	var deleteTerm int

	flag.IntVar(&deleteTerm, "term", 43200, "set delete term minute. default value is 30days")
	flag.StringVar(&confFile, "conf", "config.tml", "config toml file")

	flag.Parse()

	utils.RandSeed()

	config.Init(confFile)
	conf := config.LoadConfig()

	DeleteTerm = deleteTerm

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

	ProcessHardDelete()

	elapsed := time.Since(start)
	log.Printf("Job took %s", elapsed)
}

func ProcessHardDelete() {
	// Get App
	appList, err := AppDao.GetAllList()
	if err != nil {
		log.Fatalf("Failed get AppList, Error : %v", err)
	}

	dayAgoForDelete := time.Now().Add(time.Duration(-DeleteTerm) * time.Minute)

	for _, app := range appList {
		// Get Versions that Soft Deleted
		versionList, err := VersionDao.GetSoftDeletedVersionByAppId(app.AppId, dayAgoForDelete)
		if err != nil {
			log.Printf("Failed get Version of App %v, Error : %v", app.AppId, err)
			continue
		}
		err = deleteVersion(versionList)
		if err != nil {
			log.Printf("Failed Delete Versions of App %v, Error : %v", app.AppId, err)
			continue
		}
	}
}

func deleteVersion(versionList []models.Version) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}

	for _, version := range versionList {

		// Do Hard Delete file that Soft Deleted.
		err := FileDao.HardDelete(version.AppId, version.VersionId, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		err = FileUrlDao.HardDelete(version.AppId, version.VersionId, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		err = ResourceDao.HardDelete(version.AppId, version.VersionId, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		err = ResourceUrlDao.HardDelete(version.AppId, version.VersionId, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		// Do Hard Delete version that Soft Deleted.
		err = VersionDao.HardDelete(version.AppId, version.VersionId, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}
