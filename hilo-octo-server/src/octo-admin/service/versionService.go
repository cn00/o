package service

import (
	"database/sql"
	"strconv"

	"octo/models"

	"github.com/pkg/errors"
)

type VersionService struct {
}

func (*VersionService) GetVersion(appId int, versionId int) (models.Version, error) {

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return models.Version{}, err
	}

	return version, nil
}

func (*VersionService) GetVersions(appId int) ([]models.Version, error) {

	versions, err := versionDao.GetListByAppIds(appId)
	if err != nil {
		return []models.Version{}, err
	}

	return versions, nil
}

func (*VersionService) UpdateVersion(appId int, versionId int, description string, copyVersionId string, copyAppId string, envId int, apiAesKey string) error {

	copyVersion := sql.NullInt64{Int64: 0, Valid: false}
	if copyVersionId != "" {
		copyVersionIdInt, err := strconv.Atoi(copyVersionId)
		if err != nil {
			return errors.Wrap(err, "failed strconv.AtoI")
		}
		copyVersion = sql.NullInt64{Int64: int64(copyVersionIdInt), Valid: true}
	}

	copyApp := sql.NullInt64{Int64: 0, Valid: false}
	if copyAppId != "" {
		copyAppIdInt, err := strconv.Atoi(copyAppId)
		if err != nil {
			return errors.Wrap(err, "failed strconv.Atoi")
		}
		copyApp = sql.NullInt64{Int64: int64(copyAppIdInt), Valid: true}
	}
	err := versionDao.Update(appId, versionId, description, copyVersion, copyApp, envId, apiAesKey)
	if err != nil {
		return errors.Wrap(err, "failed to Update Version.Description")
	}
	return nil
}

func (*VersionService) CheckDestinationVersionId(appId int, sourceVersionId int, destinationVersionId int) error {
	version, err := versionDao.Get(appId, sourceVersionId)
	if err != nil {
		return err
	}
	if version.CopyVersionId.Valid && int(version.CopyVersionId.Int64) != destinationVersionId {
		return nil
	}

	destinationVersion, err := versionDao.Get(appId, destinationVersionId)
	if err != nil {
		return err
	}
	if version.EnvId.Valid && destinationVersion.EnvId.Valid && version.EnvId.Int64 == destinationVersion.EnvId.Int64 {
		return nil
	}
	return errors.New("copyVersionId != destinationVersionId or Env ID is not same")
}

func (*VersionService) DeleteVersion(appId, versionId int) error {
	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	err = fileDao.DeleteAllByVersionId(appId, versionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to Delete File")
	}

	err = fileUrlDao.DeleteAllByVersionId(appId, versionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to Delete File Url")
	}

	err = resourceDao.DeleteAllByVersionId(appId, versionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to Delete Resource")
	}

	err = resourceUrlDao.DeleteAllByVersionId(appId, versionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to Delete Resource Url")
	}

	err = versionDao.Delete(appId, versionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to Delete Version")
	}
	return tx.Commit()
}
