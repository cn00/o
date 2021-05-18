package controllers

import (
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

func CopyEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)
	switch c.Param("type") {
	case "ab":
		copyAssetBundleEndpoint(c, app)
	case "r":
		copyResourceEndpoint(c, app)
	case "appab":
		copyAppAssetBundleEndpoint(c, app)
	case "appr":
		copyAppResourceEndpoint(c, app)
	default:
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func copyAssetBundleEndpoint(c *gin.Context, app models.App) {
	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
		DryRun               bool     `json:"dry_run"`
	}
	if c.BindJSON(&json) != nil {
		return
	}

	res, err := fileCopyService.CopySelectedFile(filecopyservice.CopySelectedFileOptions{
		AppId:                app.AppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
		DryRun:               json.DryRun,
	}, conf.Api.EnvCheck)
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

func copyAppAssetBundleEndpoint(c *gin.Context, app models.App) {
	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationAppId     int      `json:"destination_app_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
		DryRun               bool     `json:"dry_run"`
	}

	if c.BindJSON(&json) != nil {
		return
	}

	res, err := fileCopyService.CopySelectedFile(filecopyservice.CopySelectedFileOptions{
		AppId:                app.AppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationAppId:     json.DestinationAppId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
		DryRun:               json.DryRun,
	}, conf.Api.EnvCheck)

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

func copyResourceEndpoint(c *gin.Context, app models.App) {
	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
		DryRun               bool     `json:"dry_run"`
	}
	if c.BindJSON(&json) != nil {
		return
	}

	res, err := resourceCopyService.CopySelectedFile(resourcecopyservice.CopySelectedFileOptions{
		AppId:                app.AppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
		DryRun:               json.DryRun,
	}, conf.Api.EnvCheck)
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

func copyAppResourceEndpoint(c *gin.Context, app models.App) {
	var json struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationAppId     int      `json:"destination_app_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
		DryRun               bool     `json:"dry_run"`
	}
	if c.BindJSON(&json) != nil {
		return
	}

	res, err := resourceCopyService.CopySelectedFile(resourcecopyservice.CopySelectedFileOptions{
		AppId:                app.AppId,
		DestinationAppId:     json.DestinationAppId,
		SourceVersionId:      json.SourceVersionId,
		DestinationVersionId: json.DestinationVersionId,
		Filenames:            json.Filenames,
		DryRun:               json.DryRun,
	}, conf.Api.EnvCheck)
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
