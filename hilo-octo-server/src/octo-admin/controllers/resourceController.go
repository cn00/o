package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"octo-admin/service"
	"octo/models"
	"octo/utils"

	"github.com/gin-gonic/gin"
)

var resourceService = service.NewResourceService()

func resourceListEndpoint(c *gin.Context) {

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
	md5Param := c.Query("md5")
	objectNameParam := c.Query("objectName")

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

	app, version, fileList, pagination, err := resourceService.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tagParam, ids, revisionIds, overRevisionId, fromDateParm, toDateParm, showDeleted, page, limit)
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

	c.HTML(http.StatusOK, "resourceList.tmpl", gin.H{
		"Title":      fmt.Sprintf("Resources - %s - %s - OCTO", version.Description, app.AppName),
		"User":       c.MustGet("User"),
		"version":    version,
		"copyVersion": copyVersion,
		"fileList":   fileList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/v/", versionId, "/resource?", url.Values{
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

func resourceDetailEndpoint(c *gin.Context) {

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

	app, version, file, fileUrl, err := resourceService.GetDetail(appId, versionId, fileId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "resourceDetail.tmpl", gin.H{
		"Title":      fmt.Sprintf("%s - Resources - %s - %s - OCTO", file.Filename, version.Description, app.AppName),
		"app":        app,
		"version":    version,
		"User":       c.MustGet("User"),
		"file":       file,
		"fileExt":    filepath.Ext(file.Filename),
		"fileUrl":    fileUrl,
		"adminFlg":   userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin),
		"updaterFlg": userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
	})
}

func resourceUpdateEndpoint(c *gin.Context) {

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
	err = resourceService.Update(appId, version, fileId, priority, tagForm)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "updated")
}

func resourceDeleteEndpoint(c *gin.Context) {

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

	err = resourceService.Delete(appId, version, fileId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "deleted")
}

func resourceDeleteSelectedFileEndpoint(c *gin.Context) {
	appId, version, fileIds, isHard, err := parseDeleteSelectedFilesParams(c)
	if err != nil {
		return
	}

	if isHard {
		err = resourceService.HardDeleteSelectedFile(appId, version, fileIds)
	} else {
		err = resourceService.DeleteSelectedFile(appId, version, fileIds)
	}

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%d/v/%d/resource", appId, version))
}

func resourceDiffEndpoint(c *gin.Context) {

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

		fileList, pagination, err = resourceService.GetDiff(appId, versionId, targetAppId, targetVersionId, nameParam, fromDateParm, toDateParm, showDeleted, page, limit)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.HTML(http.StatusOK, "resourceDiff.tmpl", gin.H{
		"Title":      fmt.Sprintf("Resource Diff - %s - %s - OCTO", version.Description, app.AppName),
		"User":       c.MustGet("User"),
		"version":    version,
		"copyVersion": targetVersion,
		"fileList":   fileList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/v/", versionId, "/resourceDiff?", url.Values{
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

func resourceDiffFileEndpoint(c *gin.Context) {
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

	nameParam := c.Query("name")

	fileIdParam := c.Param("fileid")
	fileId, err := strconv.Atoi(fileIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, fileIdParam+" is not fileId")
		return
	}

	targetAppVersionIdParam := c.Query("targetAppVersionId")
	if targetAppVersionIdParam == "" {
		c.String(http.StatusBadRequest, targetAppVersionIdParam+" is not targetAppVersionId")
		return
	}

	targetAppId, targetVersionId, err := parseTargetAppVersionID(targetAppVersionIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if !userService.CheckAuthority(c, targetAppId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, _, resource, resourceUrl, err := resourceService.GetDetail(appId, versionId, fileId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_, _, _, tresourceUrl, err := resourceService.GetDetailByName(targetAppId, targetVersionId, resource.Filename)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "resourceDiffFile.tmpl", gin.H{
		"Title":      fmt.Sprintf("Resource Diff File - %s - %s - OCTO", version.Description, app.AppName),
		"User":       c.MustGet("User"),
		"noHeader":   true,
		"version":    version,
		"appId":                   appId,
		"app":                     app,
		"versionId":               versionId,
		"nameParam":               nameParam,
		"targetAppVersionIdParam": targetAppVersionIdParam,
		"adminFlg":                userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
		"file":                    resource,
		"fileUrl":             resourceUrl,
		"tfileUrl":            tresourceUrl,
	})
}

func resourceCsvOutputEndpoint(c *gin.Context) {

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
	md5Param := c.Query("md5")
	objectNameParam := c.Query("objectName")

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

	// GetList假值fileList全部取得
	page := 1
	limit := 100

	_, _, fileList, _, err := resourceService.GetList(appId, versionId, nameParam, objectNameParam, md5Param, tagParam, ids, revisionIds, overRevisionId, fromDateParm, toDateParm, showDeleted, page, limit)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	fileName := fmt.Sprintf("resource_%v_%v.csv", versionId, utils.GetDate("2006-01-02_15-04"))
	resourceService.OutputCsv(fileList, fileName, c.Writer, c.Request)
}

func resourceDiffCsvOutputEndpoint(c *gin.Context) {

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

	showDeleted := func() bool {
		if c.Query("_showdeleted") != "" {
			return c.Query("showdeleted") == "true"
		}
		return true
	}()

	// GetDiff假值fileList全部取得
	page := 1
	limit := 100

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

		fileList, _, err = resourceService.GetDiff(appId, versionId, targetAppId, targetVersionId, nameParam, fromDateParm, toDateParm, showDeleted, page, limit)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	fileName := fmt.Sprintf("resource_diff_%v-%v_%v.csv", versionId, targetVersionId, utils.GetDate("2006-01-02_15-04"))

	resourceService.OutputDiffCsv(fileList, fileName, c.Writer, c.Request)
}
