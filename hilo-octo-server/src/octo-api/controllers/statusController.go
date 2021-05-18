package controllers

import (
	"net/http"

	"octo-api/config"

	"github.com/gin-gonic/gin"
)

func StatusReadOnlyEndpoint(c *gin.Context) {
	conf := config.LoadConfig()
	c.JSON(http.StatusOK, conf.Api.ReadOnly)
}
