package controllers

import (
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"github.com/gin-gonic/gin"
)

var deleteService = &service.DeleteService{}

func DeleteAssetBundleEndpoint(c *gin.Context) {
	var res struct {
		Error string
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

	var json service.Json
	if c.BindJSON(&json) == nil {

		err := deleteService.DeleteAssetBundle(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func DeleteResourceEndpoint(c *gin.Context) {
	var res struct {
		Error string
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

	var json service.Json
	if c.BindJSON(&json) == nil {

		err := deleteService.DeleteResource(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}
