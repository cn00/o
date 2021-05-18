package service

import (
	"database/sql"
	"github.com/pkg/errors"
	"octo/models"
	"octo/utils"
)

type ItemService struct {
	itemDao    *models.AdminItemDao
	itemUrlDao *models.ItemUrlDao
}

func (s *ItemService) Update(appId int, versionId int, fileId int, priority int, tagForm string) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.update(appId, versionId, fileId, priority, tagForm, tx))
}

func (s *ItemService) update(appId int, versionId int, fileId int, priority int, tagForm string, tx *sql.Tx) error {
	tagArray := utils.SplitTags(tagForm)
	for _, tag := range tagArray {
		err := tagDao.AddTag(appId, tag, tx)
		if err != nil {
			return err
		}
	}
	return s.itemDao.Update(appId, versionId, fileId, priority, tagForm, tx)
}

func (s *ItemService) Delete(appId int, versionId int, fileId int) error {
	return s.itemDao.Delete(appId, versionId, fileId)
}

func (s *ItemService) DeleteSelectedFile(appId int, versionId int, fileIds []int) error {
	return s.itemDao.DeleteByList(appId, versionId, fileIds)
}

func (s *ItemService) HardDeleteSelectedFile(appId int, versionId int, fileIds []int) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}

	for _, id := range fileIds {
		var item models.Item
		item, err = s.itemDao.GetItemById(appId, versionId, id)
		if err != nil {
			break
		}

		err = s.itemDao.HardDelete(appId, versionId, id, tx)
		if err != nil {
			break
		}

		err = s.itemUrlDao.HardDeleteByObjectName(appId, versionId, item.ObjectName.String, tx)
		if err != nil {
			break
		}
	}

	return models.FinishTransaction(tx, errors.Wrap(err, "exec error"))
}
