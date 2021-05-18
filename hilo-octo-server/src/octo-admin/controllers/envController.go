package controllers

import (
	"fmt"
	"net/http"

	"strconv"

	"octo/models"
	"octo/utils"

	"net/url"

	"database/sql"

	"octo/service/envservice"

	"github.com/gin-gonic/gin"
)

var (
	paginationUtil = &utils.PaginationUtil{}
	envService     = envservice.NewEnvService()
)

func envListEndPoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)
	appId, err := getParamAppId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	app, err := appService.GetApp(appId)

	envList, err := envService.GetEnvList(appId)

	pageQuery := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		c.String(http.StatusBadRequest, pageQuery+" is not page")
		return
	}
	limitQuery := c.DefaultQuery("limit", "15")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		c.String(http.StatusBadRequest, limitQuery+" is not limit")
		return
	}

	pagination, err := paginationUtil.GetPagenation(envList, page, limit)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tmplFile := "envList.tmpl"

	c.HTML(http.StatusOK, tmplFile, gin.H{
		"Title":      fmt.Sprintf("Env List - %s - OCTO", app.AppName),
		"appId":      appId,
		"App":        app,
		"User":       user,
		"envList":    envList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/env?", url.Values{
			"limit": []string{limitQuery},
		}.Encode(), "&"),
		"adminFlg": userService.CheckAuthority(c, appId, models.UserRoleTypeUser),
	})
}

func envAddEndPoint(c *gin.Context) {
	user := c.MustGet("User").(models.User)
	appId, err := getParamAppId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}
	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	app, err := appService.GetApp(appId)
	c.HTML(http.StatusOK, "envAdd.tmpl", gin.H{
		"Title": fmt.Sprintf("Add Env - %s - OCTO", app.AppName),
		"appId": appId,
		"App":   app,
		"User":  user,
	})
}

func envCreateEndPoint(c *gin.Context) {
	appId, err := getParamAppId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeUser) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var env models.Env
	name := c.PostForm("name")
	detail := c.PostForm("detail")
	env.AppId = appId
	env.Name = name
	env.Detail = sql.NullString{String: detail}

	err = envService.CreateEnv(env)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%d/env", appId))

}

//func envDeleteEndPoint(c *gin.Context) {
//	appId, err := getParamAppId(c)
//	if err != nil {
//		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
//		return
//	}
//
//	if !userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin) {
//		c.AbortWithStatus(http.StatusUnauthorized)
//		return
//	}
//
//	envId := c.PostForm("envId")
//	if len(envId) == 0 {
//		c.AbortWithError(http.StatusBadRequest, errors.New("Env Id is Null"))
//		return
//	}
//
//	envIdInt, err := strconv.Atoi(envId)
//	if err != nil {
//		c.AbortWithError(http.StatusInternalServerError, err)
//		return
//	}
//	err = envService.DeleteEnv(envIdInt, appId)
//	if err != nil {
//		c.AbortWithError(http.StatusInternalServerError, err)
//		return
//	}
//	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%d/env", appId))
//}
