package service

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"octo-api/cache"
	"octo-api/config"
	"octo-api/metrics"
	"octo/models"
	"octo/utils"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"hilo-octo-proto/go/octo"
)

var (
	emptyFile        = models.File{}
	emptyFileUrl     = models.FileUrl{}
	emptyResource    = models.Resource{}
	emptyResourceUrl = models.ResourceUrl{}
)

var (
	listCache = &cache.ListCache{}
	urlCache  = &cache.UrlCache{}
)

var gcsBucketObjectRegexp = regexp.MustCompile(`/b/([^/]*)/o/(.*)`)

type DownloadService struct {
}

func (s *DownloadService) List(appId int, versionId int, revisionId int) ([]byte, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, err
	}

	var bucket models.Bucket
	err = bucketDao.GetBucket(&bucket, appId)
	if err != nil {
		return nil, err
	}

	cacheEnabled := false ;// listAPICacheEnabled(appId)
	if cacheEnabled {
		if b, exist := listCache.ListGet(appId, versionId, revisionId, maxRevisionId); exist {
			metrics.MemoryCacheHit.Add(1)
			return b, nil
		}
		metrics.MemoryCacheMiss.Add(1)
	}

	tagList, err := tagDao.GetList(appId)
	if err != nil {
		return nil, err
	}
	tagMap := map[string]int{}
	var tagNameList []string
	for _, t := range tagList {
		tagMap[t.Name] = t.TagId
		tagNameList = append(tagNameList, t.Name)
	}

	Database := new(octo.Database)

	fileList, err := fileDao.GetList(appId, versionId, revisionId)
	if err != nil {
		return nil, err
	}
	for _, f := range fileList {
		if f.RevisionId > maxRevisionId {
			return nil, errors.Wrapf(errors.New("assetbundle max revisionid err"), f.ObjectName.String)
		}
		data := new(octo.Data)
		data.Id = proto.Int(f.Id)
		data.Name = proto.String(f.Filename)
		data.Size = proto.Int(f.Size)
		data.Crc = proto.Uint32(f.Crc)
		data.Md5 = proto.String(f.Md5.String)
		data.Priority = proto.Int(f.Priority)

		tags := utils.SplitTags(f.Tag.String)
		var tagIds []int32
		for _, tag := range tags {
			tagId, ok := tagMap[tag]
			if !ok {
				err := errors.Errorf("tag not found: %v", tag)
				return nil, errors.Wrapf(err, "assetbundle object %v error", f.ObjectName.String)
			}
			tagIds = append(tagIds, int32(tagId))
		}
		if f.Generation.Valid {
			data.Generation = proto.Uint64(uint64(f.Generation.Int64))
		} else {
			data.Generation = proto.Uint64(0)
		}
		data.Tagid = tagIds
		data.State = getDataState(f.State)
		data.UploadVersionId = proto.Int(int(f.UploadVersionId.Int64))

		var dependencyIds []int32
		for _, did := range strings.Split(f.Dependency.String, ",") {
			if did != "" {
				didInt, err := strconv.Atoi(did)
				if err != nil {
					return nil, errors.Wrapf(err, "assetbundle object %v error", f.ObjectName.String)
				}
				dependencyIds = append(dependencyIds, int32(didInt))
			}
		}
		data.Dependencie = dependencyIds
		data.ObjectName = proto.String(f.ObjectName.String)
		Database.AssetBundleList = append(Database.AssetBundleList, data)
	}

	resourceList, err := resourceDao.GetList(appId, versionId, revisionId)
	if err != nil {
		return nil, err
	}
	for _, resource := range resourceList {
		if resource.RevisionId > maxRevisionId {
			return nil, errors.Wrapf(errors.New("resource max revisionid err"), resource.ObjectName.String)
		}
		data := new(octo.Data)
		data.Id = proto.Int(resource.Id)
		data.Name = proto.String(resource.Filename)
		data.Size = proto.Int(resource.Size)
		data.Md5 = proto.String(resource.Md5.String)
		data.Priority = proto.Int(resource.Priority)

		tags := utils.SplitTags(resource.Tag.String)
		var tagIds []int32
		for _, tag := range tags {
			tagId, ok := tagMap[tag]
			if !ok {
				err := errors.Errorf("tag not found: %v", tag)
				return nil, errors.Wrapf(err, "resource object %v error", resource.ObjectName.String)
			}
			tagIds = append(tagIds, int32(tagId))
		}
		if resource.Generation.Valid {
			data.Generation = proto.Uint64(uint64(resource.Generation.Int64))
		} else {
			data.Generation = proto.Uint64(0)
		}
		data.UploadVersionId = proto.Int(int(resource.UploadVersionId.Int64))
		data.Tagid = tagIds
		data.State = getDataState(resource.State)
		data.ObjectName = proto.String(resource.ObjectName.String)

		Database.ResourceList = append(Database.ResourceList, data)
	}

	Database.Revision = proto.Int(maxRevisionId)
	Database.Tagname = tagNameList
	Database.UrlFormat = proto.String(s.cdnUrl(appId) + "/" + bucket.BucketName + "-{v}-{type}/{o}?generation={g}")
	database, err := proto.Marshal(Database)
	if err != nil {
		return nil, errors.Wrap(err, "marshal error")
	}

	json, err := json.MarshalIndent(Database, "  ", "  ")
	ioutil.WriteFile("database.json", json, os.ModePerm)

	if cacheEnabled {
		listCache.ListSet(appId, versionId, revisionId, maxRevisionId, database)
	}

	return database, nil
}

func (s *DownloadService) ListAssetBundleWithAssets(appId, versionId int) (utils.List, error) {
	fileList, err := fileDao.GetList(appId, versionId, 0)
	if err != nil {
		return nil, err
	}

	assetList := utils.List{}
	for _, f := range fileList {
		assetList = append(assetList, struct{
			Assetbundle string   `json:"assetbundle"`
			Assets      []string `json:"assets"`
		}{f.Filename, f.GetAssets()})
	}

	return assetList, nil
}

func (s *DownloadService) GetAssetBundleUrl(appId int, versionId int, revisionId int, objectName string) ([]byte, string, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, "", err
	}
	u, err := s.FindAssetBundleUrl(appId, versionId, revisionId, objectName, maxRevisionId, s.cdnUrl(appId))
	if err != nil {
		return nil, "", err
	}
	res, err := proto.Marshal(&u)
	return res, u.GetUrl(), err
}

func (s *DownloadService) GetResourceUrl(appId int, versionId int, revisionId int, objectName string) ([]byte, string, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, "", err
	}
	u, err := s.FindResourceUrl(appId, versionId, revisionId, objectName, maxRevisionId, s.cdnUrl(appId))
	if err != nil {
		return nil, "", err
	}
	res, err := proto.Marshal(&u)
	return res, u.GetUrl(), errors.Wrap(err, "proto.Marshal error")
}

func (s *DownloadService) GetAssetBundleUrlList(appId int, versionId int, revisionId int, objectNameList []string) ([]byte, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, err
	}
	cdnUrl := s.cdnUrl(appId)
	var uList []*octo.Url
	for _, objectName := range objectNameList {
		u, err := s.FindAssetBundleUrl(appId, versionId, revisionId, objectName, maxRevisionId, cdnUrl)
		if err != nil {
			return nil, err
		}
		uList = append(uList, &u)
	}
	list := &octo.UrlList{
		Url: uList,
	}
	res, err := proto.Marshal(list)
	return res, errors.Wrap(err, "proto.Marshal error")
}

func (s *DownloadService) GetAssetBundleUrlListByName(appId int, versionId int, revisionId int, nameList []string) ([]*octo.Url, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, err
	}
	cdnUrl := s.cdnUrl(appId)
	var uList []*octo.Url
	for _, objectName := range nameList {
		u, err := s.FindAssetBundleUrlByName(appId, versionId, revisionId, objectName, maxRevisionId, cdnUrl)
		if err != nil {
			uList = append(uList, nil)
			continue
		}
		uList = append(uList, &u)
	}
	return uList, nil
}

func (s *DownloadService) GetResourceUrlList(appId int, versionId int, revisionId int, objectNameList []string) ([]byte, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, err
	}
	cdnUrl := s.cdnUrl(appId)
	var uList []*octo.Url
	for _, objectName := range objectNameList {
		u, err := s.FindResourceUrl(appId, versionId, revisionId, objectName, maxRevisionId, cdnUrl)
		if err != nil {
			return nil, err
		}
		uList = append(uList, &u)
	}
	list := &octo.UrlList{
		Url: uList,
	}
	res, err := proto.Marshal(list)
	return res, errors.Wrap(err, "proto.Marshal error")
}

func (s *DownloadService) GetResourceUrlListByName(appId int, versionId int, revisionId int, nameList []string) ([]*octo.Url, error) {
	maxRevisionId, err := versionDao.GetMaxRevision(appId, versionId)
	if err != nil {
		return nil, err
	}
	cdnUrl := s.cdnUrl(appId)
	var uList []*octo.Url
	for _, objectName := range nameList {
		u, err := s.FindResourceUrlByName(appId, versionId, revisionId, objectName, maxRevisionId, cdnUrl)
		if err != nil {
			uList = append(uList, nil)
			continue
		}
		uList = append(uList, &u)
	}
	return uList, nil
}

func (s *DownloadService) FindAssetBundleUrl(appId int, versionId int, revisionId int, objectName string, maxRevisionId int, cdnUrl string) (octo.Url, error) {
	file, err := fileDao.GetByObjectName(appId, versionId, objectName)
	if err != nil {
		return octo.Url{}, err
	}
	if file == emptyFile {
		err := &ObjectNotFoundError{
			AppId:      appId,
			VersionId:  versionId,
			ObjectName: objectName,
		}
		return octo.Url{}, errors.Wrap(err, "file not found")
	}

	var u octo.Url
	if revisionId >= file.RevisionId {
		url, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, file.RevisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyFileUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: file.RevisionId,
				ObjectName: objectName,
			}
			return octo.Url{}, errors.Wrap(err, "file url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_LATEST.Enum()
	} else {
		url, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, revisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyFileUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: revisionId,
				ObjectName: objectName,
			}
			return octo.Url{}, errors.Wrap(err, "file url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_OLD.Enum()
	}
	return u, nil
}

func (s *DownloadService) FindAssetBundleUrlByName(appId int, versionId int, revisionId int, name string, maxRevisionId int, cdnUrl string) (octo.Url, error) {
	if u, exist := urlCache.AssetBundleGet(appId, versionId, revisionId, name, maxRevisionId); exist {
		metrics.MemoryCacheHit.Add(1)
		return u, nil
	}

	metrics.MemoryCacheMiss.Add(1)

	file, err := fileDao.GetByName(appId, versionId, name)
	if err != nil {
		return octo.Url{}, err
	}
	if file == emptyFile {
		err := &ObjectNotFoundError{
			AppId:      appId,
			VersionId:  versionId,
			Name:       name,
		}
		return octo.Url{}, errors.Wrap(err, "file not found")
	}

	var u octo.Url
	objectName := file.ObjectName.String
	if revisionId >= file.RevisionId {
		url, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, file.RevisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyFileUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: file.RevisionId,
				Name:       name,
			}
			return octo.Url{}, errors.Wrap(err, "file url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_LATEST.Enum()
	} else {
		url, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, revisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyFileUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: revisionId,
				Name:       name,
			}
			return octo.Url{}, errors.Wrap(err, "file url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_OLD.Enum()
	}
	urlCache.AssetBundleSet(appId, versionId, revisionId, name, maxRevisionId, u)
	return u, nil
}

func (s *DownloadService) FindResourceUrl(appId int, versionId int, revisionId int, objectName string, maxRevisionId int, cdnUrl string) (octo.Url, error) {
	resource, err := resourceDao.GetByObjectName(appId, versionId, objectName)
	if err != nil {
		return octo.Url{}, err
	}
	if resource == emptyResource {
		err := &ObjectNotFoundError{
			AppId:      appId,
			VersionId:  versionId,
			ObjectName: objectName,
		}
		return octo.Url{}, errors.Wrap(err, "resource not found")
	}

	var u octo.Url
	if revisionId >= resource.RevisionId {
		url, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, resource.RevisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyResourceUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: resource.RevisionId,
				ObjectName: objectName,
			}
			return octo.Url{}, errors.Wrap(err, "resource url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_LATEST.Enum()
	} else {
		url, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, revisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyResourceUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: revisionId,
				ObjectName: objectName,
			}
			return octo.Url{}, errors.Wrap(err, "resource url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_OLD.Enum()
	}
	return u, nil
}

func (s *DownloadService) FindResourceUrlByName(appId int, versionId int, revisionId int, name string, maxRevisionId int, cdnUrl string) (octo.Url, error) {
	if u, exist := urlCache.ResourceGet(appId, versionId, revisionId, name, maxRevisionId); exist {
		metrics.MemoryCacheHit.Add(1)
		return u, nil
	}

	metrics.MemoryCacheMiss.Add(1)

	resource, err := resourceDao.GetByName(appId, versionId, name)
	if err != nil {
		return octo.Url{}, err
	}
	if resource == emptyResource {
		err := &ObjectNotFoundError{
			AppId:      appId,
			VersionId:  versionId,
			Name:       name,
		}
		return octo.Url{}, errors.Wrap(err, "resource not found")
	}

	var u octo.Url
	objectName := resource.ObjectName.String
	if revisionId >= resource.RevisionId {
		url, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, resource.RevisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyResourceUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: resource.RevisionId,
				Name:       name,
			}
			return octo.Url{}, errors.Wrap(err, "resource url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_LATEST.Enum()
	} else {
		url, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, objectName, revisionId)
		if err != nil {
			return octo.Url{}, err
		}
		if url == emptyResourceUrl {
			err := &ObjectNotFoundError{
				AppId:      appId,
				VersionId:  versionId,
				RevisionId: revisionId,
				Name:       name,
			}
			return octo.Url{}, errors.Wrap(err, "resource url not found")
		}
		curl, err := s.createUrl(url.Url, cdnUrl)
		if err != nil {
			return octo.Url{}, err
		}
		u.Revision = proto.Int(url.RevisionId)
		u.Url = proto.String(curl)
		u.State = octo.Url_OLD.Enum()
	}
	urlCache.ResourceSet(appId, versionId, revisionId, objectName, maxRevisionId, u)
	return u, nil
}

func (s *DownloadService) GetMaxRevision(appId int, versionId int) ([]byte, error) {
	revisionId, err := s.GetMaxRevisionInt(appId, versionId)
	if err != nil {
		return nil, err
	}

	db := new(octo.Database)
	db.Revision = proto.Int(revisionId)
	return proto.Marshal(db)
}

func (*DownloadService) GetMaxRevisionInt(appId int, versionId int) (int, error) {
	return versionDao.GetMaxRevision(appId, versionId)
}

func (*DownloadService) GetVersion(appId int, versionId int) (models.Version, error) {
	return versionDao.Get(appId, versionId)
}

func (*DownloadService) createUrl(url, cdnUrl string) (string, error) {
	if url == "" {
		return "", errors.New("url is empty")
	}
	m := gcsBucketObjectRegexp.FindStringSubmatch(url)
	if len(m) == 0 {
		return "", errors.New("match error")
	}
	s := cdnUrl + "/" + m[1] + "/" + m[2]
	return s, nil
}

func (*DownloadService) cdnUrl(appId int) string {
	cdnConfig := config.LoadConfig().CDN
	if s := cdnConfig.Apps[strconv.Itoa(appId)]; s != "" {
		return s
	}
	return cdnConfig.Default
}
