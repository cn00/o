package controllers

import (
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var authService = &service.AuthService{}

type AuthMiddleware struct{}

func (m *AuthMiddleware) ClientAuthV1(c *gin.Context) {
	app, err := m.appByClientSecretKey(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	//if (app == models.App{}) {
	//	abortWithInvalidSecretKeyError(c)
	//	return
	//}
	c.Set("app", app)
}

func (m *AuthMiddleware) AppAuthV1(c *gin.Context) {
	app, err := m.appByAppSecretKey(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	//if (app == models.App{}) {
	//	abortWithInvalidSecretKeyError(c)
	//	return
	//}
	c.Set("app", app)
}

func (m *AuthMiddleware) ClientAuthV2(c *gin.Context) {
	app, err := m.appByClientSecretKey(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if (app == models.App{}) || !m.validateParamAppId(c, app) {
		abortWithInvalidSecretKeyError(c)
		return
	}
	c.Set("app", app)
}

func (m *AuthMiddleware) AppAuthV2(c *gin.Context) {
	app, err := m.appByAppSecretKey(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	//if (app == models.App{}) || !m.validateParamAppId(c, app) {
	//	abortWithInvalidSecretKeyError(c)
	//	return
	//}
	c.Set("app", app)
}

func (m *AuthMiddleware) appByClientSecretKey(c *gin.Context) (models.App, error) {
	key := m.headerSecretKey(c)
	return authService.GetAppByClientSecretKey(key)
}

func (m *AuthMiddleware) appByAppSecretKey(c *gin.Context) (models.App, error) {
	key := m.headerSecretKey(c)
	return authService.GetAppByAppSecretKey(key)
}

func (m *AuthMiddleware) headerSecretKey(c *gin.Context) string {
	return c.Request.Header.Get("X-Octo-Key")
}

func (m *AuthMiddleware) validateParamAppId(c *gin.Context, app models.App) bool {
	appId, err := strconv.Atoi(c.Param("appId"))
	if err != nil {
		return false
	}
	return app.AppId == appId
}

func abortWithInvalidSecretKeyError(c *gin.Context) {
	c.AbortWithError(http.StatusUnauthorized,
		errors.New("Invalid SecretKey")).SetType(gin.ErrorTypePublic)
}
