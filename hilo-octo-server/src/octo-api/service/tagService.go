package service

import (
	"log"

	"database/sql"
	"octo/models"
	"octo/utils"
)

var tagService = &TagService{}

type TagService struct{}

func (s *TagService) UpdateAssetBundle(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.updateAssetBundle(app, versionId, json, tx))
}

func (s *TagService) updateAssetBundle(app models.App, versionId int, json Json, tx *sql.Tx) error {

	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	if err := s.ensureTags(app.AppId, json.Tags, tx); err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		//get file record
		file, err := fileDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.File{}) {
			log.Println(fileName, "not exist")
			continue
		}

		tags := utils.SplitTags(file.Tag.String)
		newTags := utils.MergeTags(tags, json.Tags)
		newTagsString := utils.JoinTags(newTags)
		if newTagsString == file.Tag.String {
			continue
		}

		err = fileDao.UpdateTag(app.AppId, versionId, file.Id, revision, newTagsString, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TagService) UpdateResource(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.updateResource(app, versionId, json, tx))
}

func (s *TagService) updateResource(app models.App, versionId int, json Json, tx *sql.Tx) error {

	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	if err := s.ensureTags(app.AppId, json.Tags, tx); err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		//get file record
		file, err := resourceDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.Resource{}) {
			log.Println(fileName, "not exist")
			continue
		}

		tags := utils.SplitTags(file.Tag.String)
		newTags := utils.MergeTags(tags, json.Tags)
		newTagsString := utils.JoinTags(newTags)
		if newTagsString == file.Tag.String {
			continue
		}

		err = resourceDao.UpdateTag(app.AppId, versionId, file.Id, revision, newTagsString, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (*TagService) ensureTags(appId int, tags []string, tx *sql.Tx) error {
	for _, tag := range tags {
		if err := tagDao.AddTag(appId, tag, tx); err != nil {
			return err
		}
	}
	return nil
}

func (s *TagService) RemoveAssetBundle(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.removeAssetBundle(app, versionId, json, tx))
}

func (s *TagService) removeAssetBundle(app models.App, versionId int, json Json, tx *sql.Tx) error {

	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		//get file record
		file, err := fileDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.File{}) {
			log.Println(fileName, "not exist")
			continue
		}

		tags := utils.SplitTags(file.Tag.String)

		var newTagsString string
		if len(json.Tags) != 0 {
			newTags := utils.RemoveTags(tags, json.Tags)
			newTagsString = utils.JoinTags(newTags)
		}

		if newTagsString == file.Tag.String {
			continue
		}

		err = fileDao.UpdateTag(app.AppId, versionId, file.Id, revision, newTagsString, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TagService) RemoveResource(app models.App, versionId int, json Json) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.removeResource(app, versionId, json, tx))
}

func (s *TagService) removeResource(app models.App, versionId int, json Json, tx *sql.Tx) error {

	revision, err := versionDao.IncrementMaxRevision(app.AppId, versionId, tx)
	if err != nil {
		return err
	}

	//for name
	for _, fileName := range json.Files {
		//get file record
		file, err := resourceDao.GetByName(app.AppId, versionId, fileName)
		if err != nil {
			return err
		}
		if (file == models.Resource{}) {
			log.Println(fileName, "not exist")
			continue
		}

		tags := utils.SplitTags(file.Tag.String)

		var newTagsString string
		if len(json.Tags) != 0 {
			newTags := utils.RemoveTags(tags, json.Tags)
			newTagsString = utils.JoinTags(newTags)
		}

		if newTagsString == file.Tag.String {
			continue
		}

		err = resourceDao.UpdateTag(app.AppId, versionId, file.Id, revision, newTagsString, tx)
		if err != nil {
			return err
		}
	}

	return nil
}
