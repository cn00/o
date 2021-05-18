package service

import (
	"octo/models"
	"octo/utils"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

type CheckService struct{}

type ActiveAssetBundleResponse struct {
	Filename    string   `json:"filename"`
	Tag         []string `json:"tag"`
	Id          int      `json:"id"`
	Revision    int      `json:"revision"`
	Size        int      `json:"size"`
	Version     int      `json:"version"`
	Crc         uint32   `json:"crc"`
	Md5         string   `json:"md5"`
	Priority    int      `json:"priority"`
	Dependency  string   `json:"dependency"`
	UpdDatetime string   `json:"upd_datetime"`
}

type ActiveResourceResponse struct {
	Filename    string   `json:"filename"`
	Tag         []string `json:"tag"`
	Id          int      `json:"id"`
	Revision    int      `json:"revision"`
	Size        int      `json:"size"`
	Version     int      `json:"version"`
	Md5         string   `json:"md5"`
	Priority    int      `json:"priority"`
	UpdDatetime string   `json:"upd_datetime"`
}

type DiffAssetBundleResponse struct {
	File    models.File
	FileUrl models.FileUrl
}

type DiffResourceResponse struct {
	File    models.Resource
	FileUrl models.ResourceUrl
}

func (service *CheckService) ListActiveAssetBundle(appId int, versionId int) ([]ActiveAssetBundleResponse, error) {
	tagMap, err := service.getTagMap(appId)
	if err != nil {
		return nil, err
	}

	fileList, err := fileDao.GetList(appId, versionId, 0)
	if err != nil {
		return nil, err
	}

	list := make([]ActiveAssetBundleResponse, 0, len(fileList))
	for _, f := range fileList {
		state := *getDataState(f.State)
		if state != octo.Data_ADD && state != octo.Data_UPDATE {
			continue
		}

		tags := utils.SplitTags(f.Tag.String)
		if err := service.validateTag(tags, tagMap); err != nil {
			return nil, err
		}

		list = append(list, ActiveAssetBundleResponse{
			Filename:    f.Filename,
			Tag:         tags,
			Id:          f.Id,
			Revision:    f.RevisionId,
			Size:        f.Size,
			Version:     f.VersionId,
			Crc:         f.Crc,
			Md5:         f.Md5.String,
			Priority:    f.Priority,
			Dependency:  f.Dependency.String,
			UpdDatetime: f.UpdDatetime.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil
}

func (service *CheckService) ListRangeActiveAssetBundle(appId int, versionId int, startDate string, endDate string) ([]ActiveAssetBundleResponse, error) {
	tagMap, err := service.getTagMap(appId)
	if err != nil {
		return nil, err
	}
	fileList, err := fileDao.GetRangeList(appId, versionId, 0, startDate, endDate)
	if err != nil {
		return nil, err
	}

	list := make([]ActiveAssetBundleResponse, 0, len(fileList))
	for _, f := range fileList {
		state := *getDataState(f.State)
		if state != octo.Data_ADD && state != octo.Data_UPDATE {
			continue
		}

		tags := utils.SplitTags(f.Tag.String)
		if err := service.validateTag(tags, tagMap); err != nil {
			return nil, err
		}

		list = append(list, ActiveAssetBundleResponse{
			Filename:    f.Filename,
			Tag:         tags,
			Id:          f.Id,
			Revision:    f.RevisionId,
			Size:        f.Size,
			Version:     f.VersionId,
			Crc:         f.Crc,
			Md5:         f.Md5.String,
			Priority:    f.Priority,
			Dependency:  f.Dependency.String,
			UpdDatetime: f.UpdDatetime.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil

}

func (service *CheckService) ListActiveResource(appId int, versionId int) ([]ActiveResourceResponse, error) {
	tagMap, err := service.getTagMap(appId)
	if err != nil {
		return nil, err
	}

	fileList, err := resourceDao.GetList(appId, versionId, 0)
	if err != nil {
		return nil, err
	}

	list := make([]ActiveResourceResponse, 0, len(fileList))
	for _, f := range fileList {
		state := *getDataState(f.State)
		if state != octo.Data_ADD && state != octo.Data_UPDATE {
			continue
		}

		tag := utils.SplitTags(f.Tag.String)
		if err := service.validateTag(tag, tagMap); err != nil {
			return nil, err
		}

		list = append(list, ActiveResourceResponse{
			Filename:    f.Filename,
			Tag:         tag,
			Id:          f.Id,
			Revision:    f.RevisionId,
			Size:        f.Size,
			Version:     f.VersionId,
			Md5:         f.Md5.String,
			Priority:    f.Priority,
			UpdDatetime: f.UpdDatetime.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil
}

func (service *CheckService) getTagMap(appId int) (map[string]bool, error) {
	tagList, err := tagDao.GetList(appId)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]bool, len(tagList))
	for _, t := range tagList {
		tagMap[t.Name] = true
	}

	return tagMap, nil
}

func (service *CheckService) validateTag(tag []string, tagMap map[string]bool) error {
	for _, t := range tag {
		if !tagMap[t] {
			return errors.Errorf("tag not found: %v", tag)
		}
	}
	return nil
}

func (service *CheckService) DiffAssetBundle(appId int, versionId int, revisionId int, targetRevisionId int) ([]DiffAssetBundleResponse, error) {
	var res []DiffAssetBundleResponse
	fileList, err := fileDao.GetDiffList(appId, versionId, revisionId, targetRevisionId)
	if err != nil {
		return res, errors.Wrap(err, "failed to get diff list")
	}
	for _, file := range fileList {
		fileUrl, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, file.ObjectName.String, file.RevisionId)
		if err != nil {
			return res, errors.Wrap(err, "failed to get diff list url")
		}
		diffAssetBundleResponse := DiffAssetBundleResponse{File: file, FileUrl: fileUrl}
		res = append(res, diffAssetBundleResponse)
	}
	return res, nil
}

func (service *CheckService) DiffResource(appId int, versionId int, revisionId int, targetRevisionId int) ([]DiffResourceResponse, error) {
	var res []DiffResourceResponse
	fileList, err := resourceDao.GetDiffList(appId, versionId, revisionId, targetRevisionId)
	if err != nil {
		return res, errors.Wrap(err, "failed to get diff list")
	}
	for _, file := range fileList {
		fileUrl, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, file.ObjectName.String, file.RevisionId)
		if err != nil {
			return res, errors.Wrap(err, "failed to get diff list url")
		}
		diffResourceResponse := DiffResourceResponse{File: file, FileUrl: fileUrl}
		res = append(res, diffResourceResponse)
	}
	return res, nil
}
