package service

import (
	"database/sql"
	"log"

	"octo/models"
)

type DeleteService struct {
}

type Json struct {
	Files []string
	Tags  []string
}

func (s *DeleteService) DeleteAssetBundle(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.deleteAssetBundle(app, versionId, json, tx))
}

func (*DeleteService) deleteAssetBundle(app models.App, versionId int, json Json, tx *sql.Tx) error {
	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		log.Println(fileName, json.Tags)
		//get file record
		file, err := fileDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.File{}) {
			log.Println(fileName, "not exist")
			continue
		}

		err = fileDao.Delete(app.AppId, versionId, file.Id, revision, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DeleteService) DeleteResource(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.deleteResource(app, versionId, json, tx))
}

func (*DeleteService) deleteResource(app models.App, versionId int, json Json, tx *sql.Tx) error {
	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		log.Println(fileName, json.Tags)
		//get file record
		file, err := resourceDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.Resource{}) {
			log.Println(fileName, "not exist")
			continue
		}

		err = resourceDao.Delete(app.AppId, versionId, file.Id, revision, tx)
		if err != nil {
			return err
		}
	}

	return nil
}
