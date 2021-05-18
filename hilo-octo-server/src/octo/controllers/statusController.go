package controllers

import (
	"net/http"

	"octo/models"

	"github.com/gin-gonic/gin"
)

func StatusEndpoint(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func StatusDBEndpoint(c *gin.Context) {
	for _, err := range []error{
		models.CheckDBM(),
		models.CheckDBS(),
	} {
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err).SetType(gin.ErrorTypePublic)
			return
		}
	}
	c.String(http.StatusOK, "ok")
}
