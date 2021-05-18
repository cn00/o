package controllers

import (
	"errors"
	"net/http"

	"octo/models"
	"octo/service/filecopyservice"
	"octo/service/resourcecopyservice"

	"github.com/gin-gonic/gin"
)

var (
	fileCopyService     = filecopyservice.NewFileCopyService()
	resourceCopyService = resourcecopyservice.NewResourceCopyService()
)

func assetBundleCopySelectedFileEndpoint(c *gin.Context) {
	appId, err := getParamAppId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationAppId     int      `json:"destination_app_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
	}
	if c.BindJSON(&json) != nil {
		return
	}

	if json.DestinationAppId > 0 && json.DestinationVersionId == 0 {
		c.AbortWithError(http.StatusInternalServerError, errors.New("copyVersionId is null"))
		return
	}

	if err := versionService.CheckDestinationVersionId(appId, json.SourceVersionId, json.DestinationVersionId); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res, err := fileCopyService.CopySelectedFile(filecopyservice.CopySelectedFileOptions{
		AppId:                appId,
		DestinationAppId:     json.DestinationAppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
	}, conf.Admin.EnvCheck)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, struct {
		Result map[string]string `json:"result"`
	}{
		Result: res,
	})
}

func resourceCopySelectedFileEndpoint(c *gin.Context) {
	appId, err := getParamAppId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationAppId     int      `json:"destination_app_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
	}
	if c.BindJSON(&json) != nil {
		return
	}

	if err := versionService.CheckDestinationVersionId(appId, json.SourceVersionId, json.DestinationVersionId); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if err := envService.CheckSameEnvironment(appId, json.SourceVersionId, json.DestinationAppId, json.DestinationVersionId,
		conf.Admin.EnvCheck); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res, err := resourceCopyService.CopySelectedFile(resourcecopyservice.CopySelectedFileOptions{
		AppId:                appId,
		DestinationAppId:     json.DestinationAppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
	}, conf.Admin.EnvCheck)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, struct {
		Result map[string]string `json:"result"`
	}{
		Result: res,
	})
}
