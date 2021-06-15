package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"octo-admin/service"
	"octo/models"
	"octo/utils"

	"github.com/gin-gonic/gin"
)

var (
	userService = &service.UserService{}
	fileService = service.NewFileService()
)

func assetBundleListEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	nameParam := c.Query("name")
	tagParam := c.Query("tag")
	idsParam := c.Query("ids")
	objectNameParam := c.Query("objectName")
	md5Param := c.Query("md5")
	var ids []int
	if idsParam != "" {
		for _, idStr := range strings.Split(idsParam, ",") {
			id, _ := strconv.Atoi(idStr)
			ids = append(ids, id)
		}
	}
	revisionIdsParam := c.Query("revisionids")
	revisionIds, overRevisionId := utils.GetSearchRange(revisionIdsParam)

	fromDateParm := c.Query("fromdate")
	toDateParm := c.Query("todate")
	err, errMsg := utils.CheckFromToDateFormat(fromDateParm, toDateParm)
	if err != nil {
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	showDeleted := func() bool {
		if c.Query("_showdeleted") != "" {
			return c.Query("showdeleted") == "true"
		}
		return false
	}()

	pageQuery := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		errMsg := pageQuery + " in not limit"
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}
	limitQuery := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		errMsg := limitQuery + " in not limit"
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	app, version, fileList, pagination, err := fileService.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tagParam, ids, revisionIds, overRevisionId, fromDateParm, toDateParm, showDeleted, page, limit)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var copyVersion models.Version
	if version.CopyAppId.Valid && version.CopyVersionId.Valid {
		copyVersion, err = versionService.GetVersion(int(version.CopyAppId.Int64), int(version.CopyVersionId.Int64))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.HTML(http.StatusOK, "fileList.tmpl", gin.H{
		"Title":      fmt.Sprintf("AssetBundle - %s - %s - OCTO", version.Description, app.AppName),
		"User":       c.MustGet("User"),
		"version":    version,
		"copyVersion": copyVersion,
		"fileList":   fileList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/v/", versionId, "/file?", url.Values{
			"name":         []string{nameParam},
			"objectName":   []string{objectNameParam},
			"md5":          []string{md5Param},
			"tag":          []string{tagParam},
			"ids":          []string{idsParam},
			"revisionids":  []string{revisionIdsParam},
			"showdeleted":  []string{fmt.Sprint(showDeleted)},
			"_showdeleted": []string{"on"},
			"limit":        []string{limitQuery},
		}.Encode(), "&"),
		"appId":            appId,
		"app":              app,
		"versionId":        versionId,
		"nameParam":        nameParam,
		"objectNameParam":  objectNameParam,
		"md5Param":         md5Param,
		"tagParam":         tagParam,
		"idsParam":         idsParam,
		"revisionIdsParam": revisionIdsParam,
		"showDeleted":      showDeleted,
		"adminFlg":         userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
		"limit":            limit,
		"fromDateParam":    fromDateParm,
		"toDateParam":      toDateParm,
	})
}

func assetBundleDetailEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}
	fileIdParam := c.Param("fileid")
	fileId, err := strconv.Atoi(fileIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, fileIdParam+" is not fileId")
		return
	}

	app, version, file, fileUrl, err := fileService.GetDetail(appId, versionId, fileId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "fileDetail.tmpl", gin.H{
		"Title":      fmt.Sprintf("%s - AssetBundle - %s - %s - OCTO", file.Filename, version.Description, app.AppName),
		"app":        app,
		"version":    version,
		"User":       c.MustGet("User"),
		"file":       file,
		"fileUrl":    fileUrl,
		"adminFlg":   userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin),
		"updaterFlg": userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
	})
}

func assetBundleUpdateEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}
	fileIdParam := c.Param("fileid")
	fileId, err := strconv.Atoi(fileIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, fileIdParam+" is not fileId")
		return
	}

	tagForm := c.PostForm("tag")

	priority := 0
	err = fileService.Update(appId, version, fileId, priority, tagForm)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "updated")
}

func assetBundleDeleteEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}
	fileIdParam := c.Param("fileid")
	fileId, err := strconv.Atoi(fileIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, fileIdParam+" is not fileId")
		return
	}

	err = fileService.Delete(appId, version, fileId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "deleted")
}

func assetBundleDeleteSelectedFileEndpoint(c *gin.Context) {
	appId, version, fileIds, isHard, err := parseDeleteSelectedFilesParams(c)
	if err != nil {
		return
	}

	if isHard {
		err = fileService.HardDeleteSelectedFile(appId, version, fileIds)
	} else {
		err = fileService.DeleteSelectedFile(appId, version, fileIds)
	}
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%d/v/%d/file", appId, version))
}

func assetBundleDiffEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	version, err := versionService.GetVersion(appId, versionId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	app, err := appService.GetApp(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	appVersions, err := versionService.GetVersions(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	nameParam := c.Query("name")

	showDeleted := func() bool {
		if c.Query("_showdeleted") != "" {
			return c.Query("showdeleted") == "true"
		}
		return true
	}()

	var fileList utils.List
	var pagination utils.Pagination

	targetAppVersionIdParam := c.Query("targetAppVersionId")

	if targetAppVersionIdParam == "" && version.CopyVersionId.Valid && version.CopyAppId.Valid {
		targetAppVersionIdParam = fmt.Sprint(version.CopyAppId.Int64, "_", version.CopyVersionId.Int64)
	}

	fromDateParm := c.Query("fromdate")
	toDateParm := c.Query("todate")

	err, errMsg := utils.CheckFromToDateFormat(fromDateParm, toDateParm)
	if err != nil {
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	pageQuery := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		errMsg := pageQuery + " in not limit"
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}
	limitQuery := c.DefaultQuery("limit", "500")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		errMsg := limitQuery + " in not limit"
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	var targetVersion models.Version
	if targetAppVersionIdParam != "" {
		targetAppId, targetVersionId, err := parseTargetAppVersionID(targetAppVersionIdParam)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		if !userService.CheckAuthority(c, targetAppId, models.UserRoleTypeReader) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		targetVersion, err = versionService.GetVersion(targetAppId, targetVersionId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		fileList, pagination, err = fileService.GetDiff(appId, versionId, targetAppId, targetVersionId, nameParam, fromDateParm, toDateParm, showDeleted, page, limit)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.HTML(http.StatusOK, "fileDiff.tmpl", gin.H{
		"Title":      fmt.Sprintf("AssetBundle Diff - %s - %s - OCTO", version.Description, app.AppName),
		"User":       c.MustGet("User"),
		"version":    version,
		"copyVersion": targetVersion,
		"fileList":   fileList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/v/", versionId, "/fileDiff?", url.Values{
			"showdeleted":        []string{fmt.Sprint(showDeleted)},
			"_showdeleted":       []string{"on"},
			"targetAppVersionId": []string{targetAppVersionIdParam},
			"limit":              []string{fmt.Sprint(limit)},
		}.Encode(), "&"),
		"appId":                   appId,
		"app":                     app,
		"appVersions":             appVersions,
		"versionId":               versionId,
		"nameParam":               nameParam,
		"showDeleted":             showDeleted,
		"targetAppVersionIdParam": targetAppVersionIdParam,
		"adminFlg":                userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
		"limit":                   limit,
		"fromDateParam":           fromDateParm,
		"toDateParam":             toDateParm,
	})

}

func parseTargetAppVersionID(s string) (int, int, error) {
	appVersionIDSlice := strings.Split(s, "_")
	if len(appVersionIDSlice) != 2 {
		return 0, 0, errors.New("malformed target app version id")
	}

	appIDStr, versionIDStr := appVersionIDSlice[0], appVersionIDSlice[1]
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return 0, 0, errors.New(appIDStr + " is not appId")
	}

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		return 0, 0, errors.New(versionIDStr + " is not version")
	}

	return appID, versionID, nil
}

func assetBundleCsvOutputEndpoint(c *gin.Context) {
	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	nameParam := c.Query("name")
	tagParam := c.Query("tag")
	idsParam := c.Query("ids")
	objectNameParam := c.Query("objectName")
	md5Param := c.Query("md5")
	var ids []int
	if idsParam != "" {
		for _, idStr := range strings.Split(idsParam, ",") {
			id, _ := strconv.Atoi(idStr)
			ids = append(ids, id)
		}
	}
	revisionIdsParam := c.Query("revisionids")
	revisionIds, overRevisionId := utils.GetSearchRange(revisionIdsParam)

	fromDateParm := c.Query("fromdate")
	toDateParm := c.Query("todate")
	err, errMsg := utils.CheckFromToDateFormat(fromDateParm, toDateParm)
	if err != nil {
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}
	log.Println(c.Query("_showdeleted"), c.Query("showdeleted"))
	showDeleted := func() bool {
		if c.Query("_showdeleted") != "" {
			return c.Query("showdeleted") == "true"
		}
		return false
	}()

	// GetList假值fileList全部取得
	page := 1
	limit := 100

	_, _, fileList, _, err := fileService.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tagParam, ids, revisionIds, overRevisionId, fromDateParm, toDateParm, showDeleted, page, limit)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fileName := fmt.Sprintf("assetBundle_%v_%v.csv", versionId, utils.GetDate("2006-01-02_15-04"))
	fileService.OutputCsv(fileList, fileName, c.Writer, c.Request)
}

func assetBundleDiffCsvOutputEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	version, err := versionService.GetVersion(appId, versionId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	nameParam := c.Query("name")

	showDeleted := func() bool {
		if c.Query("_showdeleted") != "" {
			return c.Query("showdeleted") == "true"
		}
		return true
	}()

	var fileList utils.List
	// var pagination utils.Pagination

	targetAppVersionIdParam := c.Query("targetAppVersionId")

	if targetAppVersionIdParam == "" && version.CopyVersionId.Valid && version.CopyAppId.Valid {
		targetAppVersionIdParam = fmt.Sprint(version.CopyAppId.Int64, "_", version.CopyVersionId.Int64)
	}

	fromDateParm := c.Query("fromdate")
	toDateParm := c.Query("todate")

	err, errMsg := utils.CheckFromToDateFormat(fromDateParm, toDateParm)
	if err != nil {
		log.Println(errMsg)
		c.String(http.StatusBadRequest, errMsg)
		return
	}

	// GetDiff假值fileList全部取得
	page := 1
	limit := 100

	// var targetVersion models.Version
	var targetAppId int
	var targetVersionId int
	if targetAppVersionIdParam != "" {
		targetAppId, targetVersionId, err = parseTargetAppVersionID(targetAppVersionIdParam)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		if !userService.CheckAuthority(c, targetAppId, models.UserRoleTypeReader) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		fileList, _, err = fileService.GetDiff(appId, versionId, targetAppId, targetVersionId, nameParam, fromDateParm, toDateParm, showDeleted, page, limit)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	fileName := fmt.Sprintf("assetBundle_diff_%v-%v_%v.csv", versionId, targetVersionId, utils.GetDate("2006-01-02_15-04"))
	fileService.OutputDiffCsv(fileList, fileName, c.Writer, c.Request)
}
