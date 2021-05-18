package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"octo/models"
	"strconv"
	"strings"
)

func parseDeleteSelectedFilesParams(c *gin.Context) (appId int, versionId int, fileIds []int, isHard bool, err error) {
	appIdParam := c.Param("appid")
	appId, err = strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	versionParam := c.Param("version")
	versionId, err = strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}
	fileIdsParam := c.PostForm("fileids")
	fileIdStrings := strings.Split(fileIdsParam, ",")
	for _, fileId := range fileIdStrings {
		var id int
		id, err = strconv.Atoi(fileId)
		if err != nil {
			c.String(http.StatusBadRequest, fileIdsParam+" is not fileId")
			return
		}
		fileIds = append(fileIds, id)
	}

	isHardParam := c.PostForm("isHard")
	if isHardParam != "" {
		isHard, err = strconv.ParseBool(isHardParam)
		if err != nil {
			c.String(http.StatusBadRequest, isHardParam+" is not boolean")
			return
		}
	}
	return
}
