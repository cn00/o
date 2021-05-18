package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func getParamAppId(c *gin.Context) (int, error) {
	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		return 0, errors.New(appIdParam + " is not appId")
	}
	return appId, nil
}
