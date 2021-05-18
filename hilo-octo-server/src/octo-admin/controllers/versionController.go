package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"octo/models"

	"github.com/gin-gonic/gin"
)

func versionDetailEndpoint(c *gin.Context) {

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

	app, err := appService.GetApp(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	version, err := versionService.GetVersion(appId, versionId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	envList, err := envService.GetEnvList(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	tmplFile := "versionEdit.tmpl"
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		tmplFile = "versionDetail.tmpl"
	}
	c.HTML(http.StatusOK, tmplFile, gin.H{
		"Title":   fmt.Sprintf("Version %d Details - %s - OCTO", versionId, app.AppName),
		"App":     app,
		"Version": version,
		"EnvList": envList,
		"User":    c.MustGet("User"),
	})
}

func versionUpdateEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	descriptionForm := c.PostForm("description")
	if descriptionForm == "" {
		c.String(http.StatusBadRequest, "description is empty")
		return
	}

	copyVersionIdForm := c.PostForm("copyVersionId")
	_, err = strconv.Atoi(copyVersionIdForm)
	if copyVersionIdForm != "" && err != nil {
		c.String(http.StatusBadRequest, copyVersionIdForm+" is not copy version")
		return
	}

	copyAppIdForm := c.PostForm("copyAppId")

	// TODO ENV管理機能を追加するまで臨時
	envIdForm := c.PostForm("envId")
	var envId int
	if len(envIdForm) > 0 {
		envId, err = strconv.Atoi(envIdForm)
		if err != nil {
			c.String(http.StatusBadRequest, envIdForm+" is not appId")
			return
		}
	}

	apiAesKey := c.PostForm("apiaeskey")

	err = versionService.UpdateVersion(appId, version, descriptionForm, copyVersionIdForm, copyAppIdForm, envId, apiAesKey)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%s/v/%s/detail", appIdParam, versionParam))
}

func versionDeleteEndpoint(c *gin.Context) {
	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	err = versionService.DeleteVersion(appId, version)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprint("/"))
}
