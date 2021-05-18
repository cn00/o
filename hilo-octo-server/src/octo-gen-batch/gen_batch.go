package main

import (
	"log"
	"net/url"
	"strconv"

	"octo-api/config"
	"octo-api/service"
	"octo/models"
	"octo/utils"

	"fmt"

	"flag"
	"time"
)

var appDao = &models.AppDao{}
var fileDao = models.NewFileDao()
var fileUrlDao = &models.FileUrlDao{}
var resourceDao = models.NewResourceDao()
var resourceUrlDao = &models.ResourceUrlDao{}
var bucketDao = &models.BucketDao{}

func main() {
	log.Println("Generation Batch Start")
	start := time.Now()
	var confFile string
	var appId int
	var isAllUpdate bool
	flag.StringVar(&confFile, "conf", "config.tml", "config toml file")
	flag.IntVar(&appId, "appId", 0, "set appId. If appId is zero(0), batch run for all app ")
	flag.BoolVar(&isAllUpdate, "WarningAllUpdate", false, "Warning! If you set this flag, update all data of generation!")

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

	if !isAllUpdate {
		UpdateGenerationAndUploadVersionIdOfNullData(appId)
	} else {
		UpdateGenerationFromUrlsWithApp(appId)
	}

	elapsed := time.Since(start)
	log.Printf("Job took %s", elapsed)
}

func UpdateGenerationAndUploadVersionIdOfNullData(appId int) error {
	log.Println("Start update generation that null data.")
	apps, err := getApps(appId)
	if err != nil {
		log.Fatal(err)
		return err
	}
	for _, app := range apps {
		log.Printf("App : %s, AppId: %d Start\n", app.AppName, app.AppId)
		err := updateGenerationAndUploadVersionIdOfNullData(app.AppId)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateGenerationAndUploadVersionIdOfNullData(appId int) error {
	fmt.Println("File Start")

	files, err := fileDao.GetByGenIsNullOrUploadVerIdNull(appId)
	if err != nil {
		return err
	}
	log.Printf("Generation is Null of Files Count: %d\n", len(files))
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	for _, file := range files {
		fileUrl, err := fileUrlDao.GetUrlByObjectNameAndRevisionId(file.AppId, file.VersionId, file.ObjectName.String, file.RevisionId)
		if err != nil {
			log.Printf("fileUrl Get Error : %v", err)
			continue
		}
		if len(fileUrl.Url) > 0 {
			genInt, uploadVersionIdInt, err := getGenerationValue(fileUrl.Url, file.AppId)
			if err != nil {
				log.Printf("fileUrl getGenerationValue Error : %v", err)
				continue
			}

			// File更新
			err = fileDao.UpdateGenerationAndUploadVersionId(file.AppId, file.VersionId, file.RevisionId, file.Crc, fileUrl.ObjectName, genInt, uploadVersionIdInt, tx)
			if err != nil {
				log.Printf("appId : %d, versionId: %d, objectName : %s, revisionId : %d fileUrl update failed \n", file.AppId, file.VersionId, file.ObjectName.String, file.RevisionId)
				log.Printf("fileUrl UpdateGenerationAndUploadVersionId Error : %v", err)
				tx.Rollback()
				return err
			}
		} else {
			log.Printf("appId : %d, versionId: %d, objectName : %s, revisionId : %d fileUrl is none \n", file.AppId, file.VersionId, file.ObjectName.String, file.RevisionId)
		}

	}
	tx.Commit()

	resources, err := resourceDao.GetByGenIsNullOrUploadVerIdNull(appId)
	if err != nil {
		return err
	}

	log.Println("Resource Start")
	log.Printf("Generation is Null of Resources Count: %d\n", len(resources))
	tx2, err := models.StartTransaction()
	if err != nil {
		return err
	}

	for _, r := range resources {
		resourceUrl, err := resourceUrlDao.GetUrlByObjectNameAndRevisionId(r.AppId, r.VersionId, r.ObjectName.String, r.RevisionId)
		if err != nil {
			log.Printf("resourceUrl Get Error : %v", err)
			continue
		}
		if len(resourceUrl.Url) > 0 {
			genInt, uploadVersionIdInt, err := getGenerationValue(resourceUrl.Url, r.AppId)
			if err != nil {
				log.Printf("resourceUrl getGenerationValue Error : %v", err)
				continue
			}

			err = resourceDao.UpdateGenerationAndUploadVersionId(r.AppId, r.VersionId, r.RevisionId, resourceUrl.ObjectName, genInt, uploadVersionIdInt, tx2)
			if err != nil {
				log.Printf("appId : %d, versionId: %d, objectName : %s, revisionId : %d fileUrl update failed \n", r.AppId, r.VersionId, r.ObjectName.String, r.RevisionId)
				log.Printf("resourceUrl UpdateGenerationAndUploadVersionId Error : %v", err)
				tx2.Rollback()
				return err
			}
		} else {
			log.Printf("appId : %d, versionId: %d, objectName : %s, revisionId : %d resourceUrl is none \n", r.AppId, r.VersionId, r.ObjectName.String, r.RevisionId)
		}
	}
	tx2.Commit()
	return nil

}

func UpdateGenerationFromUrlsWithApp(appId int) error {
	log.Println("Start update generation of files for App.")

	apps, err := getApps(appId)
	if err != nil {
		log.Fatal(err)
		return err
	}
	for _, app := range apps {
		log.Printf("App : %s, AppId: %d Start\n", app.AppName, app.AppId)
		err := updateGenerationByUrls(app)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateGenerationByUrls(app models.App) error {
	fileUrls, err := fileUrlDao.GetListByAppId(app.AppId)
	if err != nil {
		return err
	}

	log.Println("File Start")
	log.Printf("File Count: %d\n", len(fileUrls))

	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}

	for _, fileUrl := range fileUrls {
		if len(fileUrl.Url) > 0 {
			genInt, uploadVersionIdInt, err := getGenerationValue(fileUrl.Url, fileUrl.AppId)
			if err != nil {
				log.Printf("updateGenerationByUrls getGenerationValue Error : %v\n", err)
				continue
			}

			file, err := fileDao.GetByObjectName(fileUrl.AppId, fileUrl.VersionId, fileUrl.ObjectName)
			if err != nil {
				log.Printf("updateGenerationByUrls GetByObjectName Error : %v\n", err)
				continue
			}

			// File更新
			err = fileDao.UpdateGenerationAndUploadVersionId(file.AppId, file.VersionId, file.RevisionId, file.Crc, fileUrl.ObjectName, genInt, uploadVersionIdInt, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			log.Printf("AppId: %d File VersionId : %d, RevisionId: %d, ObjectName: %s url is null\n", fileUrl.AppId, fileUrl.VersionId, fileUrl.RevisionId, fileUrl.ObjectName)
		}
	}
	tx.Commit()

	resourceUrls, err := resourceUrlDao.GetListByAppId(app.AppId)
	log.Println("Resource Start")
	log.Printf("Resource Count : %d\n", len(resourceUrls))
	if err != nil {
		return err
	}
	// Reousrce更新
	tx2, err := models.StartTransaction()
	if err != nil {
		return err
	}
	for _, resourceUrl := range resourceUrls {
		if len(resourceUrl.Url) > 0 {
			genInt, uploadVersionIdInt, err := getGenerationValue(resourceUrl.Url, app.AppId)

			resource, err := resourceDao.GetByObjectName(resourceUrl.AppId, resourceUrl.VersionId, resourceUrl.ObjectName)
			if err != nil {
				log.Printf("updateGenerationByUrls GetByObjectName Error : %v\n", err)
				continue
			}

			err = resourceDao.UpdateGenerationAndUploadVersionId(resource.AppId, resource.VersionId, resource.RevisionId, resourceUrl.ObjectName, genInt, uploadVersionIdInt, tx2)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			log.Printf("AppId: %d Resource VersionId : %d, RevisionId: %d, ObjectName: %s url is null\n", resourceUrl.AppId, resourceUrl.VersionId, resourceUrl.RevisionId, resourceUrl.ObjectName)
		}
	}
	tx2.Commit()
	return nil
}

func getApps(appId int) ([]models.App, error) {
	var apps []models.App
	var err error
	if appId == 0 {
		apps, err = appDao.GetAllList()
		if err != nil {
			return []models.App{}, err
		}
	} else {
		var app models.App
		err = appDao.Get(&app, appId)
		if err != nil {
			return []models.App{}, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func getGenerationValue(fileUrl string, appId int) (uint64, int, error) {
	bucket := &models.Bucket{}
	bucketDao.GetBucket(bucket, appId)
	urlString, err := url.QueryUnescape(fileUrl)
	if err != nil {
		return 0, 0, err
	}
	parseURL, err := url.Parse(urlString)
	// get upload version id
	uploadVersionIdInt, err := utils.GetUploadVersionId(parseURL, bucket.BucketName)
	if err != nil {
		return 0, 0, err
	}
	values := parseURL.Query()
	if err != nil {
		return 0, 0, err
	}
	genInt, err := strconv.ParseUint(values.Get("generation"), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return genInt, uploadVersionIdInt, nil
}
