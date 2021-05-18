package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"octo-api/config"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func checkReadOnly(c *gin.Context) {
	conf := config.LoadConfig()
	if conf.Api.ReadOnly {
		c.AbortWithError(http.StatusServiceUnavailable,
			errors.New("API is read only")).SetType(gin.ErrorTypePublic)
		return
	}
}

func checkCliVersion(c *gin.Context) {
	cliVersionStr := c.Request.Header.Get("X-Octo-Cli-Version")
	if cliVersionStr == "" {
		return
	}

	cliVersion, err := parseCliVersion(cliVersionStr)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest,
			errors.New("cli version is required")).SetType(gin.ErrorTypePublic)
		return
	}

	conf := config.LoadConfig()
	if cliVersion < conf.Api.MinimumCliVersion {
		c.AbortWithError(http.StatusBadRequest,
			errors.Errorf("cli version %v is obsolete", cliVersion)).SetType(gin.ErrorTypePublic)
		return
	}
}

func parseCliVersion(s string) (float64, error) {
	v := strings.TrimPrefix(s, "v")
	return strconv.ParseFloat(v, 64)
}
