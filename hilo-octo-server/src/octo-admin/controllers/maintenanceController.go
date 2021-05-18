package controllers

import (
	"net/http"
	"strconv"

	"octo-admin/service"

	"github.com/gin-gonic/gin"
)

var maintenanceService = &service.MaintenanceService{}

func makeDiffSqlEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	versionParam := c.Param("version")
	versionId, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}

	tappIdParam := c.Param("tappid")
	tappId, err := strconv.Atoi(tappIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, tappIdParam+" is not appId")
		return
	}

	tversionParam := c.Param("tversion")
	tversionId, err := strconv.Atoi(tversionParam)
	if err != nil {
		c.String(http.StatusBadRequest, tversionParam+" is not version")
		return
	}

	maintenanceService.MakeDiffSql(appId, versionId, tappId, tversionId)

	c.String(http.StatusOK, "OK")
}
