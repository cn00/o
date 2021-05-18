package service

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"path"
	"strconv"
	"strings"

	"octo/models"
	"octo/utils"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
)

var adminResourceDao = models.NewAdminResourceDao()

type ResourceService struct {
	ItemService
}

func NewResourceService() *ResourceService {
	return &ResourceService{
		ItemService{
			itemDao:    &adminResourceDao.AdminItemDao,
			itemUrlDao: &resourceUrlDao.ItemUrlDao,
		},
	}
}

func (*ResourceService) GetList(appId int, versionId int, nameParam string, objectNameParam string, md5Param string, tagParam string, ids []int, revisionIds []int, overRevisionId int, fromDate string, toDate string, showDeleted bool, page int, limit int) (models.App, models.Version, utils.List, utils.Pagination, error) {

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
	fileList, err := adminResourceDao.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tags, ids, revisionIds, overRevisionId, fromDate, toDate, showDeleted)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	pagination, err := paginationUtil.GetPagenation(fileList, page, limit)
	if err != nil {
		return models.App{}, models.Version{}, nil, utils.Pagination{}, err
	}

	return app, version, fileList, pagination, nil
}

func (*ResourceService) GetDetail(appId int, versionId int, fileId int) (models.App, models.Version, models.Resource, models.ResourceUrl, error) {

	var app models.App
	err := appDao.Get(&app, appId)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	file, err := adminResourceDao.GetById(appId, versionId, fileId)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	fileUrl, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, file.ObjectName.String, file.RevisionId)

	return app, version, file, fileUrl, err
}

func (*ResourceService) GetDetailByName(appId int, versionId int, filename string) (models.App, models.Version, models.Resource, models.ResourceUrl, error) {

	var app models.App
	err := appDao.Get(&app, appId)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	version, err := versionDao.Get(appId, versionId)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	file, err := adminResourceDao.GetByFileName(appId, versionId, filename)
	if err != nil {
		return models.App{}, models.Version{}, models.Resource{}, models.ResourceUrl{}, err
	}

	fileUrl, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(appId, versionId, file.ObjectName.String, file.RevisionId)

	return app, version, file, fileUrl, err
}

func (*ResourceService) GetDiff(appId int, versionId int, targetAppId int, targetVersionId int, nameParam string, fromDate string, toDate string, showDeleted bool, page int, limit int) (utils.List, utils.Pagination, error) {

	fileList, err := adminResourceDao.GetList(appId, versionId, "", "", "", nil, nil, nil, 0, fromDate, toDate, true)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	targetFileList, err := adminResourceDao.GetList(targetAppId, targetVersionId, "", "", "", nil, nil, nil, 0, fromDate, toDate, true)
	if err != nil {
		return nil, utils.Pagination{}, err
	}

	targetFileMap := map[string]models.Resource{}
	for _, data := range targetFileList {
		file := data.(models.Resource)
		targetFileMap[file.Filename] = file
	}

	names := strings.Fields(nameParam)

	var res utils.List
	for _, data := range fileList {
		file := data.(models.Resource)
		tfile, ok := targetFileMap[file.Filename]
		diff := newResourceDifference(file, tfile, ok)
		if !diff.different() {
			continue
		}
		if !matchFilename(file.Filename, names) {
			continue
		}
		if !showDeleted && file.State == int(octo.Data_DELETE) {
			continue
		}
		res = append(res, diffResources{
			Resource: file,
			Diff:     diff,
			FileExt:  path.Ext(file.Filename),
		})
	}

	pagination, err := paginationUtil.GetPagenation(res, page, limit)
	if err != nil {
		return res, pagination, err
	}

	return res, pagination, err
}

type diffResources struct {
	models.Resource
	Diff    resourceDifference
	FileExt string
}

type resourceDifference struct {
	Absent      bool
	Md5         bool
	Tag         bool
	State       bool
	BuildNumber bool
}

func newResourceDifference(file, tfile models.Resource, ok bool) resourceDifference {
	if !ok {
		return resourceDifference{Absent: true}
	}
	return resourceDifference{
		Md5:   file.Md5 != tfile.Md5,
		Tag:   file.Tag != tfile.Tag,
		State: file.State != tfile.State,
	}
}

func (d *resourceDifference) different() bool {
	return d.Absent || d.Md5 || d.Tag || d.State || d.BuildNumber
}

func (*ResourceService) OutputCsv(fileList utils.List, fileName string, w http.ResponseWriter, r *http.Request) {

	header := []string{"FileName", "ID", "revision", "size", "ObjectName", "MD5", "tag", "BuildNumber", "datetime"}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	wr.Write(header)
	wr.Flush()

	for _, f := range fileList {
		file := f.(models.Resource)
		item := file.Item

		out := []string{
			item.Filename,
			strconv.Itoa(item.Id),
			strconv.Itoa(item.RevisionId),
			strconv.Itoa(item.Size),
			item.ObjectName.String,
			item.Md5.String,
			item.Tag.String,
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

func (*ResourceService) OutputDiffCsv(fileList utils.List, fileName string, w http.ResponseWriter, r *http.Request) {

	header := []string{"FileName", "ID", "revision", "size", "MD5", "tag", "datetime"}

	b := &bytes.Buffer{}
	wr := csv.NewWriter(b)
	wr.Write(header)
	wr.Flush()

	for _, f := range fileList {
		file := f.(diffResources)
		item := file.Item

		out := []string{
			item.Filename,
			strconv.Itoa(item.Id),
			strconv.Itoa(item.RevisionId),
			strconv.Itoa(item.Size),
			item.Md5.String,
			item.Tag.String,
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
