package service

import (
	"database/sql"
	"net/url"
	"time"

	"octo/models"
	"octo/utils"

	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type ResourceService struct {
}

func (*ResourceService) UploadList(app models.App, versionId int) (map[string]File, models.Gcs, error) {

	appId := app.AppId
	var gcs models.Gcs
	if app.StorageType == 1 {
		if err := gcsDao.GetGcs(&gcs, appId); err != nil {
			return nil, models.Gcs{}, err
		}
	}

	fileList, err := resourceDao.GetList(appId, versionId, 0)
	fileMap := map[string]File{}
	for i := range fileList {
		f := fileList[i]
		fileStruct := File{}
		fileStruct.Id = f.Id
		fileStruct.MD5 = f.Md5.String
		fileStruct.State = f.State
		if f.Generation.Valid {
			fileStruct.Generation = uint64(f.Generation.Int64)
		}
		if !f.ObjectName.Valid {
			fileStruct.EncriptedName = aesEncrypt(f.Filename, app.AesKey)
		} else {
			fileStruct.EncriptedName = f.ObjectName.String
		}
		fileMap[f.Filename] = fileStruct
	}
	return fileMap, gcs, err
}

func (*ResourceService) MakeNewResource(appId int, version int, filename string) (models.Resource, error) {

	f, err := resourceDao.GetByName(appId, version, filename)
	if err != nil {
		return models.Resource{}, err
	}
	if (f != models.Resource{}) {
		return f, nil
	}
	file := models.Resource{Item: models.Item{AppId: appId, VersionId: version, Filename: filename, State: 1}}
	for i := 0; i < 10; i++ {
		objectName := makeObjectHash()
		count, err := resourceDao.CountByObjectName(objectName)
		if err != nil {
			return models.Resource{}, err
		}
		if count > 0 {
			continue
		}
		file.ObjectName.String = objectName
		file.ObjectName.Valid = true

		err = resourceDao.InsertWithId(file)
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number == 1146 {
				continue
			}
		}
		if err != nil {
			return models.Resource{}, err
		}
		return file, nil
	}
	return models.Resource{}, errors.New("MakeObjectHashRetryError")
}

func (s *ResourceService) UploadAll(app models.App, version int, json []NewFile, useOldTagFlg bool) (int, error) {
	tx, err := models.StartTransaction()
	if err != nil {
		return 0, err
	}
	revisionId, err := s.uploadAll(app, version, json, useOldTagFlg, tx)
	return revisionId, models.FinishTransaction(tx, err)
}

func (*ResourceService) uploadAll(app models.App, version int, json []NewFile, useOldTagFlg bool, tx *sql.Tx) (int, error) {
	//add version if not exist
	if err := versionDao.AddVersion(app.AppId, version); err != nil {
		return 0, err
	}
	bucket := &models.Bucket{}
	bucketDao.GetBucket(bucket, app.AppId)
	revision, err := versionDao.IncrementMaxRevision(app.AppId, version, tx)
	if err != nil {
		return 0, err
	}

	for _, newFile := range json {
		var file = models.Resource{}
		file.AppId = app.AppId
		file.VersionId = version
		file.RevisionId = revision
		file.Filename = newFile.Filename
		urlString, err := url.QueryUnescape(newFile.Url)
		if err != nil {
			return 0, errors.Wrap(err, "query unescape error")
		}
		file.Size = int(newFile.Size)
		file.Md5 = sql.NullString{String: newFile.Md5, Valid: true}
		file.Tag = sql.NullString{String: newFile.Tag, Valid: true}
		if len(newFile.BuildNumber) > 0 {
			file.BuildNumber = sql.NullString{String: newFile.BuildNumber, Valid: true}
		}
		parseURL, err := url.Parse(urlString)
		values := parseURL.Query()
		if err != nil {
			return 0, err
		}

		genInt, err := strconv.ParseUint(values.Get("generation"), 10, 64)
		if err != nil {
			return 0, err
		}
		file.Generation.Int64 = int64(genInt)

		// get upload version id
		uploadVersionIdInt, err := utils.GetUploadVersionId(parseURL, bucket.BucketName)
		if err != nil {
			return 0, err
		}
		// set upload version id
		file.UploadVersionId = sql.NullInt64{Int64: int64(uploadVersionIdInt), Valid: true}

		file, err = resourceDao.Replace(file, useOldTagFlg, tx)
		if err != nil {
			return 0, err
		}
		if file.ObjectName.Valid {
			//create urls record
			url := models.ResourceUrl{
				AppId:       file.AppId,
				VersionId:   version,
				RevisionId:  revision,
				ObjectName:  file.ObjectName.String,
				Md5:         file.Md5,
				Url:         urlString,
				UpdDatetime: time.Now()}
			err := resourceUrlDao.AddUrl(url, tx)
			if err != nil {
				return 0, err
			}
		}

		tagArray := utils.SplitTags(newFile.Tag)
		for _, tag := range tagArray {
			err := tagDao.AddTag(app.AppId, tag, tx)
			if err != nil {
				return 0, err
			}
		}
	}
	return revision, nil
}
