package controllers

import (
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"github.com/gin-gonic/gin"
)

var assetBundleService = &service.AssetBundleService{}

func UploadListEndpoint(c *gin.Context) {

	var res struct {
		ProjectId string
		Backet    string
		Location  string
		Error     string
		Files     map[string]service.File
	}

	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}

	fileMap, gcp, err := assetBundleService.UploadList(app, v)
	if err != nil {
		c.Error(err)
		res.Error = err.Error()
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res.ProjectId = gcp.ProjectId
	res.Backet = gcp.Backet
	res.Location = gcp.Location

	res.Files = fileMap

	c.JSON(http.StatusOK, res)
}

func UploadStartEndpoint(c *gin.Context) {

	var res struct {
		FileName  string
		CRC       uint32
		Id        int
		ProjectId string
		Backet    string
		Error     string
	}

	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		res.Error = "Invalid App"
		c.JSON(http.StatusForbidden, res)
		return
	}

	versionParam := c.Query("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		res.Error = versionParam + " is not version"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	appId := app.AppId
	filename := c.Query("filename")

	file, err := assetBundleService.MakeNewFile(appId, version, filename)

	if err != nil {
		c.Error(err)
		res.Error = err.Error()
		c.JSON(http.StatusInternalServerError, res)
	} else {
		res.FileName = file.ObjectName.String
		c.JSON(http.StatusOK, res)
	}

}

func UploadAllEndpoint(c *gin.Context) {
	uploadAllEndpoint(c, false)
}

func UploadAllNoTagEndpoint(c *gin.Context) {
	uploadAllEndpoint(c, true)
}

func uploadAllEndpoint(c *gin.Context, useOldTagFlg bool) {

	var res struct {
		RevisionId int
		Error      string
	}

	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		res.Error = "Invalid App"
		c.JSON(http.StatusForbidden, res)
		return
	}

	versionParam := c.Param("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		res.Error = versionParam + " is not version"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var json []service.NewFile
	if c.BindJSON(&json) == nil {
		revisionId, err := assetBundleService.UploadAll(app, version, json, useOldTagFlg)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		res.RevisionId = revisionId
		c.JSON(http.StatusOK, res)
	}
}
