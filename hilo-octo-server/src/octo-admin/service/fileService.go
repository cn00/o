package service

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"octo/models"
	"octo/utils"

	"hilo-octo-proto/go/octo"
)

var adminFileDao = models.NewAdminFileDao()
var tagDao = &models.TagDao{}

var paginationUtil = &utils.PaginationUtil{}

type FileService struct {
	ItemService
}

func NewFileService() *FileService {
	return &FileService{
		ItemService{
			itemDao:    &adminFileDao.AdminItemDao,
			itemUrlDao: &fileUrlDao.ItemUrlDao,
		},
	}
}

func (*FileService) GetList(appId int, versionId int, nameParam string, objectNameParam string, md5Param string, tagParam string, ids []int, revisionIds []int, overRevisionId int, fromDate string, toDate string, showDeleted bool, page int, limit int) (models.App, models.Version, utils.List, utils.Pagination, error) {

	var app models.App
	err := appDao.Get(&app, appId)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	tags := utils.SplitTags(tagParam)
	fileList, err := adminFileDao.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tags, ids, revisionIds, overRevisionId, fromDate, toDate, showDeleted)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	pagination, err := paginationUtil.GetPagenation(fileList, page, limit)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	return app, version, fileList, pagination, nil

}

func (*FileService) GetDetail(appId int, versionId int, fileId int) (models.App, models.Version, models.File, models.FileUrl, error) {
	var app models.App
	err := appDao.Get(&app, appId)
	if err != nil {
		return models.App{}, models.Version{}, models.File{}, models.FileUrl{}, err
	}

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return models.App{}, models.Version{}, models.File{}, models.FileUrl{}, err
	}

	file, err := adminFileDao.GetById(appId, versionId, fileId)
	if err != nil {
		return models.App{}, models.Version{}, models.File{}, models.FileUrl{}, err
	}

	fileUrl, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, file.ObjectName.String, file.RevisionId)

	return app, version, file, fileUrl, err
}

func (s *FileService) HardDeleteSelectedFile(appId int, versionId int, fileIds []int) error {
	files, err := adminFileDao.GetList(appId, versionId, "", "", "", nil, nil, nil, -1, "", "", true)
	if err != nil {
		return err
	}

	fileIdsMap := make(map[int]bool, len(fileIds))
	for _, id := range fileIds {
		fileIdsMap[id] = true
	}

	// 不删除删除对象AssetBundle检查是否依赖于
	for _, f := range files {
		file := f.(models.File)
		if utils.IsDependent(fileIds, file.Id) {
			// 因为是删除对象，所以通过
			continue
		}
		if !file.Dependency.Valid {
			// 因为没有依赖，所以忽略
			continue
		}
		deps, err := utils.SplitDependencies(file.Dependency.String)
		if err != nil {
			return err
		}
		for _, d := range deps {
			if fileIdsMap[d] {
				return errors.Errorf("id=%d is dependent by %s", d, file.Filename)
			}
		}
	}

	return s.ItemService.HardDeleteSelectedFile(appId, versionId, fileIds)
}

func (*FileService) GetDiff(appId int, versionId int, targetAppId int, targetVersionId int, nameParam string, fromDate string, toDate string, showDeleted bool, page int, limit int) (utils.List, utils.Pagination, error) {

	fileList, err := adminFileDao.GetList(appId, versionId, "", "", "", nil, nil, nil, 0, fromDate, toDate, true)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	targetFileList, err := adminFileDao.GetList(targetAppId, targetVersionId, "", "", "", nil, nil, nil, 0, fromDate, toDate, true)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	targetFileMap := map[string]models.File{}
	for _, data := range targetFileList {
		file := data.(models.File)
		targetFileMap[file.Filename] = file
	}

	names := strings.Fields(nameParam)

	var res utils.List
	for _, data := range fileList {
		file := data.(models.File)
		tfile, ok := targetFileMap[file.Filename]
		diff := newFileDifference(file, tfile, ok)
		if !diff.different() {
			continue
		}
		if !matchFilename(file.Filename, names) {
			continue
		}
		if !showDeleted && file.State == int(octo.Data_DELETE) {
			continue
		}
		res = append(res, diffFiles{
			File: file,
			Diff: diff,
		})
	}

	pagination, err := paginationUtil.GetPagenation(res, page, limit)
	if err != nil {
		return res, pagination, err
	}

	return res, pagination, nil
}

func matchFilename(filename string, names []string) bool {
	for _, s := range names {
		if !strings.Contains(filename, s) {
			return false
		}
	}
	return true
}

type diffFiles struct {
	models.File
	Diff fileDifference
}

type fileDifference struct {
	Absent      bool
	Crc         bool
	Tag         bool
	State       bool
	BuildNumber bool
}

func newFileDifference(file, tfile models.File, ok bool) fileDifference {
	if !ok {
		return fileDifference{Absent: true}
	}
	return fileDifference{
		Crc:   file.Crc != tfile.Crc,
		Tag:   file.Tag != tfile.Tag,
		State: file.State != tfile.State,
	}
}

func (d *fileDifference) different() bool {
	return d.Absent || d.Crc || d.Tag || d.State || d.BuildNumber
}

func (*FileService) OutputCsv(fileList utils.List, fileName string, w http.ResponseWriter, r *http.Request) {

	header := []string{"AssetName", "ID", "revision", "size", "crc", "ObjectName", "MD5", "tag", "dependency", "BuildNumber", "datetime"}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	wr.Write(header)
	wr.Flush()

	for _, f := range fileList {
		file := f.(models.File)
		item := file.Item

		out := []string{
			item.Filename,
			strconv.Itoa(item.Id),
			strconv.Itoa(item.RevisionId),
			strconv.Itoa(item.Size),
			strconv.FormatInt(int64(file.Crc), 10),
			item.ObjectName.String,
			item.Md5.String,
			item.Tag.String,
			file.Dependency.String,
			item.BuildNumber.String,
			item.UpdDatetime.Format("2006/01/02 15:04:05"),
		}

		wr.Write(out)
		wr.Flush()
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	w.Write(b.Bytes())
}

func (*FileService) OutputDiffCsv(fileList utils.List, fileName string, w http.ResponseWriter, r *http.Request) {

	header := []string{"AssetName", "ID", "revision", "size", "crc", "tag", "dependency", "datetime"}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	wr.Write(header)
	wr.Flush()

	for _, f := range fileList {

		file := f.(diffFiles)
		item := file.Item

		out := []string{
			item.Filename,
			strconv.Itoa(item.Id),
			strconv.Itoa(item.RevisionId),
			strconv.Itoa(item.Size),
			strconv.FormatInt(int64(file.Crc), 10),
			item.Tag.String,
			file.Dependency.String,
			item.UpdDatetime.Format("2006/01/02 15:04:05"),
		}

		wr.Write(out)
		wr.Flush()
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	w.Write(b.Bytes())
}
