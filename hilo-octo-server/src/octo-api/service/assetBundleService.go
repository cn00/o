package service

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"octo/models"
	"octo/utils"

	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type AssetBundleService struct {
}

type File struct {
	Id              int
	CRC             uint32
	MD5             string
	EncriptedName   string
	State           int
	Generation      uint64
	BuildNumber     string
	UploadVersionId int
}

type NewFile struct {
	Crc             uint32
	Md5             string
	Filename        string
	EnciptName      string
	Size            int64
	Tag             string
	Assets          []string
	Dependency      string
	Url             string
	Generation      uint64
	BuildNumber     string
	UploadVersionId int
}

func (*AssetBundleService) UploadList(app models.App, versionId int) (map[string]File, models.Gcs, error) {

	appId := app.AppId
	//TODO GCP情報取得
	var gcs models.Gcs
	if app.StorageType == 1 {
		if err := gcsDao.GetGcs(&gcs, appId); err != nil {
			return nil, models.Gcs{}, err
		}
	}

	fileList, err := fileDao.GetList(appId, versionId, 0)
	fileMap := map[string]File{}
	for i := range fileList {
		f := fileList[i]
		fileStruct := File{}
		fileStruct.Id = f.Id
		fileStruct.CRC = f.Crc
		fileStruct.State = f.State
		fileStruct.MD5 = f.Md5.String
		if f.Generation.Valid {
			fileStruct.Generation = uint64(f.Generation.Int64)
		}
		if f.BuildNumber.Valid {
			fileStruct.BuildNumber = f.BuildNumber.String
		}
		if f.UploadVersionId.Valid {
			fileStruct.UploadVersionId = int(f.UploadVersionId.Int64)
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

func (*AssetBundleService) MakeNewFile(appId int, version int, filename string) (models.File, error) {

	f, err := fileDao.GetByName(appId, version, filename)
	if err != nil {
		return models.File{}, err
	}
	if (f != models.File{}) {
		return f, nil
	}
	file := models.File{Item: models.Item{AppId: appId, VersionId: version, Filename: filename, State: 1}}
	for i := 0; i < 10; i++ {
		objectName := makeObjectHash()
		count, err := fileDao.CountByObjectName(objectName)
		if err != nil {
			return models.File{}, err
		}
		if count > 0 {
			continue
		}
		file.ObjectName.String = objectName
		file.ObjectName.Valid = true

		err = fileDao.InsertWithId(file)
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number == 1146 {
				continue
			}
		}
		if err != nil {
			return models.File{}, err
		}
		return file, nil
	}
	return models.File{}, errors.New("MakeObjectHashRetryError")
}

func (s *AssetBundleService) UploadAll(app models.App, version int, json []NewFile, useOldTagFlg bool) (int, error) {
	tx, err := models.StartTransaction()
	if err != nil {
		return 0, err
	}
	revisionId, err := s.uploadAll(app, version, json, useOldTagFlg, tx)
	return revisionId, models.FinishTransaction(tx, err)
}

func (*AssetBundleService) uploadAll(app models.App, version int, json []NewFile, useOldTagFlg bool, tx *sql.Tx) (int, error) {

	//add version if not exist
	versionDao.AddVersion(app.AppId, version)

	revision, err := versionDao.IncrementMaxRevision(app.AppId, version, tx)
	bucket := &models.Bucket{}
	bucketDao.GetBucket(bucket, app.AppId)
	if err != nil {
		return 0, err
	}

	for _, newFile := range json {
		var file = models.File{}
		file.AppId = app.AppId
		file.VersionId = version
		file.RevisionId = revision
		file.Filename = newFile.Filename
		urlString, err := url.QueryUnescape(newFile.Url)
		if err != nil {
			err = fmt.Errorf("File %v query unescape error: %v\n", file.Filename, err)
			return 0, errors.Wrap(err, "query unescape error")
		}
		parseURL, err := url.Parse(urlString)
		values := parseURL.Query()
		if err != nil {
			err = fmt.Errorf("File %v parseURLQuery Error: %v\n", file.Filename, err)
			return 0, err
		}
		
		gen := values.Get("generation")
		if len(gen) > 0 {
			genInt, err := strconv.ParseUint(values.Get("generation"), 10, 64)
			if err != nil {
				err = fmt.Errorf("File %v ParseUint Error: %v\n", file.Filename, err)
				return 0, err
			}
			file.Generation.Int64 = int64(genInt)
		}

		file.Size = int(newFile.Size)
		file.Crc = newFile.Crc
		file.Md5 = sql.NullString{String: newFile.Md5, Valid: true}
		file.Tag = sql.NullString{String: newFile.Tag, Valid: true}

		// get upload version id
		uploadVersionIdInt, err := utils.GetUploadVersionId(parseURL, bucket.BucketName)
		if err != nil {
			err = fmt.Errorf("File %v uploadVersionIdInt Error: %v\n", file.Filename, err)
			return 0, err
		}
		// set upload version id
		file.UploadVersionId = sql.NullInt64{Int64: int64(uploadVersionIdInt), Valid: true}

		// Set BuildNumber
		if len(newFile.BuildNumber) > 0 {
			file.BuildNumber = sql.NullString{String: newFile.BuildNumber, Valid: true}
		}

		if len(newFile.Assets) > 0 {
			file.Assets = sql.NullString{String: utils.JoinAssets(newFile.Assets), Valid: true}
		}

		var dependencyIds []string
		for _, dependency := range strings.Split(newFile.Dependency, ",") {
			if dependency != "" {
				d, err := fileDao.GetByName(app.AppId, version, dependency)
				if err != nil {
					err = fmt.Errorf("File %v GetByName Error: %v\n", dependency, err)
					return 0, err
				}
				dependencyIds = append(dependencyIds, strconv.Itoa(d.Id))
			}
		}
		file.Dependency = sql.NullString{String: strings.Join(dependencyIds, ","), Valid: true}
		//if (len(dependencyIds) > 0) {
		//	log.Println("dependencyIds:", file.Filename, file.Dependency.String)
		//}

		file, err = fileDao.Replace(file, useOldTagFlg, tx)
		if err != nil {
			err = fmt.Errorf("File %v Replace Error: %v\n", file.Filename, err)
			return 0, err
		}

		if file.ObjectName.Valid {
			//create urls record
			url := models.FileUrl{
				AppId:       file.AppId,
				VersionId:   version,
				RevisionId:  revision,
				ObjectName:  file.ObjectName.String,
				Crc:         file.Crc,
				Md5:         file.Md5,
				Url:         urlString,
				UpdDatetime: time.Now(),
			}
			err := fileUrlDao.AddUrl(url, tx)
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

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func aesEncrypt(str string, keyText string) string {
	plaintext := []byte(str)

	c, err := aes.NewCipher([]byte(keyText))
	if err != nil {
		log.Printf("Error: NewCipher(%d bytes) = %s\n", len(keyText), err)
	}

	cfb := cipher.NewCFBEncrypter(c, commonIV)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)

	return hex.EncodeToString(ciphertext)
}
