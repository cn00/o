package service

import (
	"octo-admin/config"
	"octo/models"
)

var appDao = &models.AppDao{}
var versionDao = &models.VersionDao{}
var bucketDao = &models.BucketDao{}
var gcsDao = &models.GcsDao{}
var envDao = models.NewEnvDao()
var fileDao = models.NewFileDao()
var resourceDao = models.NewResourceDao()

type AppService struct {
}

type AppDetail struct {
	App      models.App
	Versions []models.Version
}

func (*AppService) GetAppDetailList(userApps models.UserApps) ([]AppDetail, error) {
	var apps []models.App
	if 0 <= userApps.GetAppIds().Position(0) {
		var err error
		apps, err = appDao.GetAllList()
		if err != nil {
			return nil, err
		}
		apps = append([]models.App{models.App{AppId: 0}}, apps...)
	} else if len(userApps.GetAppIds()) > 0 {
		var err error
		apps, err = appDao.GetListByIds(userApps.GetAppIds())
		if err != nil {
			return nil, err
		}
	}

	var appDetailList []AppDetail

	for _, app := range apps {
		versionList, err := versionDao.GetListByAppIds(app.AppId)
		if err != nil {
			return nil, err
		}
		appDetailList = append(appDetailList, AppDetail{App: app, Versions: versionList})
	}
	return appDetailList, nil
}

func (*AppService) GetApp(appId int) (models.App, error) {
	var app models.App
	if err := appDao.Get(&app, appId); err != nil {
		return models.App{}, err
	}
	return app, nil
}

func (*AppService) UpdateApp(app models.App, b models.Bucket, g models.Gcs) error {
	tx, err := models.StartTransaction()
	conf := config.LoadConfig()
	g.Location = conf.GCSProject.Location
	g.ProjectId = conf.GCSProject.ProjectId
	if err != nil {
		return err
	}
	err = appDao.UpdateApp(app, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = bucketDao.Update(b, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = gcsDao.Update(g, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return models.FinishTransaction(tx, err)
}

func (*AppService) CreateApp(app models.App) error {
	tx, err := models.StartTransaction()
	if err != nil {

		return err
	}
	err = appDao.Insert(app, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	conf := config.LoadConfig()
	b := models.Bucket{
		AppId:      app.AppId,
		BucketName: "",
	}

	err = bucketDao.Insert(b, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	g := models.Gcs{
		AppId:     app.AppId,
		Location:  conf.GCSProject.Location,
		ProjectId: conf.GCSProject.ProjectId,
		Backet:    "",
	}

	err = gcsDao.Insert(g, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return models.FinishTransaction(tx, err)
}

func (*AppService)DeleteApp(appID int) error {
	daos := []models.Dao{
		appDao,
		fileUrlDao,
		fileDao,
		gcsDao,
		resourceUrlDao,
		resourceDao,
		tagDao,
		userAppDao,
		versionDao,
		envDao,
		bucketDao,
	}

	tx, err := models.StartTransaction()
	if err != nil {

		return err
	}

	for _, dao := range daos {
		if err = models.DeleteApp(dao, appID, tx); err != nil {
			break
		}
	}

	return models.FinishTransaction(tx, err)
}
