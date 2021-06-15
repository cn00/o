package controllers

import (
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"octo/service/envservice"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var syncService = &service.SyncService{}
var versionService = &service.VersionService{}
var envService = envservice.NewEnvService()

func DiffSyncEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)
	dstVersionParam := c.Param("dstVersionId")
	dstVersion, err := strconv.Atoi(dstVersionParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(dstVersionParam+" is not srcVersion")).SetType(gin.ErrorTypePublic)
		return
	}

	srcAppIdParam := c.Param("srcAppId")
	srcAppId, err := strconv.Atoi(srcAppIdParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(srcAppIdParam+" is not srcAppId")).SetType(gin.ErrorTypePublic)
		return
	}

	srcVersionIdParam := c.Param("srcVersionId")
	srcVersion, err := strconv.Atoi(srcVersionIdParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(srcVersionIdParam+" is not srcVersion")).SetType(gin.ErrorTypePublic)
		return
	}

	err = versionService.Exists(srcAppId, srcVersion)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// V1のDiffSync是一样的App之间Sync的设想AppId指定相同的
	err = envService.CheckSameEnvironmentForSync(srcAppId, srcVersion, app.AppId, dstVersion, conf.Api.EnvCheck, true)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	revisionParam := c.Param("revisionId")
	if revisionParam == "latest" {
		err = syncService.DiffSyncLatest(app.AppId, dstVersion, srcAppId, srcVersion)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		revisionId, err := strconv.Atoi(revisionParam)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest,
				errors.New(revisionParam+" is not revisionId")).SetType(gin.ErrorTypePublic)
			return
		}

		err = syncService.DiffSync(app.AppId, dstVersion, srcAppId, srcVersion, revisionId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, nil)
}

func DiffSyncEndpointV2(c *gin.Context) {

	dstVersionParam := c.Param("dstVersionId")
	dstVersion, err := strconv.Atoi(dstVersionParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(dstVersionParam+" is not Version")).SetType(gin.ErrorTypePublic)
		return
	}

	srcAppIdParam := c.Param("srcAppId")
	srcAppId, err := strconv.Atoi(srcAppIdParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(srcAppIdParam+" is not appId")).SetType(gin.ErrorTypePublic)
		return
	}

	srcVersionIdParam := c.Param("srcVersionId")
	srcVersion, err := strconv.Atoi(srcVersionIdParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(srcVersionIdParam+" is not version")).SetType(gin.ErrorTypePublic)
		return
	}

	dstAppIdParam := c.Param("dstAppId")
	dstAppId, err := strconv.Atoi(dstAppIdParam)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New(dstAppIdParam+" is not appId")).SetType(gin.ErrorTypePublic)
		return
	}

	err = envService.CheckSameEnvironment(srcAppId, srcVersion, dstAppId, dstVersion, conf.Api.EnvCheck)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	revisionParam := c.Param("revisionId")
	if revisionParam == "latest" {
		err = syncService.DiffSyncLatest(dstAppId, dstVersion, srcAppId, srcVersion)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		revisionId, err := strconv.Atoi(revisionParam)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest,
				errors.New(revisionParam+" is not revisionId")).SetType(gin.ErrorTypePublic)
			return
		}

		err = syncService.DiffSync(dstAppId, dstVersion, srcAppId, srcVersion, revisionId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, nil)
}
