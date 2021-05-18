package service

import (
	"octo/models"

	"github.com/pkg/errors"
)

type VersionService struct {
}

func (*VersionService) Exists(appId int, versionId int) error {

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return err
	}
	if (version == models.Version{}) {
		return errors.New("Source Version does not exist")
	}
	return nil
}
