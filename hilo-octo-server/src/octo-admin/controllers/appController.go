package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"octo/models"

	"log"

	"database/sql"

	"octo/utils"

	"errors"
	"github.com/gin-gonic/gin"
)

type AppEditForm struct {
	AppId           int    `form:"appId" binding:"required"`
	AppName         string `form:"appName" binding:"required"`
	Description     string `form:"description"`
	ImageUrl        string `form:"imageUrl"`
	AppSecretKey    string `form:"appSecretKey"`
	ClientSecretKey string `form:"clientSecretKey"`
	AesKey          string `form:"aesKey"`
	Bucket          string `form:"bucket"`
	ProjectId       string `form:"projectId"`
	Location        string `form:"location"`
}

type AppNewForm struct {
	AppId       int    `form:"appId" binding:"required,min=1"`
	AppName     string `form:"appName" binding:"required"`
	Description string `form:"description"`
	ImageUrl    string `form:"imageUrl"`
	Email       string `form:"email"`
}

func appMainEndpoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)
	userApps := c.MustGet("UserApps").(models.UserApps)

	appDetailList, err := appService.GetAppDetailList(userApps)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.HTML(http.StatusOK, "main.tmpl", gin.H{
		"Title":         "Main - OCTO",
		"AppDetailList": appDetailList,
		"User":          user,
		"CheckAdminAuthority": func(appId int) bool {
			return userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin)
		},
	})
}

func appDetailEndpoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)
	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	app, err := appService.GetApp(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	bucket, err := bucketService.GetBucket(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	gcs, err := gcsService.GetGCS(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	appForm := AppEditForm{
		AppId:           app.AppId,
		AppName:         app.AppName,
		Description:     app.Description,
		ImageUrl:        app.ImageUrl.String,
		AppSecretKey:    app.AppSecretKey,
		AesKey:          app.AesKey,
		ClientSecretKey: app.ClientSecretKey,
		Bucket:          bucket.BucketName,
		Location:        gcs.Location,
		ProjectId:       gcs.ProjectId,
	}
	tmplFile := "appEdit.tmpl"
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
		tmplFile = "appDetail.tmpl"
	}
	c.HTML(http.StatusOK, tmplFile, gin.H{
		"Title":     fmt.Sprintf("App Details - %s - OCTO", app.AppName),
		"AppForm":   appForm,
		"User":      user,
		"CanDelete": userService.CheckAuthority(c, 0, models.UserRoleTypeAdmin),
	})
}

func appUpdateEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var form AppEditForm
	err = c.Bind(&form)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	app := models.App{
		AppId:       form.AppId,
		AppName:     form.AppName,
		Description: form.Description,
		ImageUrl:    sql.NullString{String: form.ImageUrl, Valid: true},
	}

	b := models.Bucket{
		AppId:      form.AppId,
		BucketName: form.Bucket,
	}

	g := models.Gcs{
		AppId:     form.AppId,
		Location:  form.Location,
		ProjectId: form.ProjectId,
		Backet:    form.Bucket,
	}
	err = appService.UpdateApp(app, b, g)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%s/detail", appIdParam))
}

func appNewEncpoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)
	if !userService.CheckAuthority(c, 0, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.HTML(http.StatusOK, "appNew.tmpl", gin.H{
		"User":  user,
		"error": "",
	})
}

func appInsertEndpoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)

	var form AppNewForm
	if err := c.Bind(&form); err != nil {

		c.HTML(http.StatusBadRequest, "appNew.tmpl", gin.H{
			"User":    user,
			"AppForm": form,
			"error":   err.Error(),
		})
		log.Print(err)
		return
	}

	existApp, err := appService.GetApp(form.AppId)

	if err == nil && existApp.AppId > 0 {
		c.HTML(http.StatusBadRequest, "appNew.tmpl", gin.H{
			"User":    user,
			"AppForm": form,
			"error":   errors.New("Already have a AppId").Error(),
		})
		log.Print(err)
		return
	}

	app := &models.App{
		AppId:           form.AppId,
		AppName:         form.AppName,
		Description:     form.Description,
		ImageUrl:        sql.NullString{String: form.ImageUrl, Valid: true},
		AppSecretKey:    utils.RandString(32),
		Email:           sql.NullString{String: form.Email, Valid: true},
		ClientSecretKey: utils.RandString(32),
		AesKey:          utils.RandString(16),
	}
	err = appService.CreateApp(*app)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func appDeleteEncpoint(c *gin.Context) {
	if !userService.CheckAuthority(c, 0, models.UserRoleTypeAdmin) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	appIDParam := c.Param("appid")
	appID, err := strconv.Atoi(appIDParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIDParam+" is not appId")
		return
	}

	err = appService.DeleteApp(appID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}
