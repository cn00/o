package controllers

import (
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"github.com/gin-gonic/gin"
)

var tagService = &service.TagService{}

func TagAssetBundleEndpoint(c *gin.Context) {
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

		err := tagService.UpdateAssetBundle(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func TagResourceEndpoint(c *gin.Context) {
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

		err := tagService.UpdateResource(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func RemoveTagAssetBundleEndpoint(c *gin.Context) {
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

		err := tagService.RemoveAssetBundle(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func RemoveTagResourceEndpoint(c *gin.Context) {
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

		err := tagService.RemoveResource(app, version, json)
		if err != nil {
			c.Error(err)
			res.Error = err.Error()
			c.JSON(http.StatusInternalServerError, res)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}
