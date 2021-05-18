package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func getParamVersionId(c *gin.Context) (int, error) {

	versionIdParam := c.Param("versionId")
	versionId, err := strconv.Atoi(versionIdParam)
	if err != nil {
		return 0, errors.New(versionIdParam + " is not version")
	}

	return versionId, nil
}
