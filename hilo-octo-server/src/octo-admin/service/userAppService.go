package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"octo/models"

	"github.com/gin-gonic/gin"
)

var userAppDao = &models.UserAppDao{}

type UserAppService struct {
}

func (*UserAppService) ListEndpoint(c *gin.Context) {

	appIdParam := c.Param("appid")
	appId, err := strconv.Atoi(appIdParam)
	if err != nil {
		c.String(http.StatusBadRequest, appIdParam+" is not appId")
		return
	}

	if !userService.CheckAuthority(c, appId, models.UserRoleTypeReader) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userList, err := userAppDao.GetListByAppId(appId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

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

	pagination, err := paginationUtil.GetPagenation(userList, page, limit)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var app models.App
	if appId != 0 {
		if err := appDao.Get(&app, appId); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.HTML(http.StatusOK, "userAppList.tmpl", gin.H{
		"Title": func() string {
			if appId == 0 {
				return "Admin Members - OCTO"
			}
			return fmt.Sprintf("Members - %s - OCTO", app.AppName)
		}(),
		"User":       c.MustGet("User"),
		"userList":   userList,
		"pagination": pagination,
		"paginationBaseUrl": fmt.Sprint("/a/", appId, "/userapp?", url.Values{
			"limit": []string{limitQuery},
		}.Encode(), "&"),
		"appId":    appId,
		"app":      app,
		"adminFlg": userService.CheckAuthority(c, appId, models.UserRoleTypeAdmin),
	})
}

func (*UserAppService) AddEndpoint(c *gin.Context) {

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

	var app models.App
	if appId != 0 {
		if err := appDao.Get(&app, appId); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.HTML(http.StatusOK, "userAppAdd.tmpl", gin.H{
		"Title": func() string {
			if appId == 0 {
				return "Add Admin - OCTO"
			}
			return fmt.Sprintf("Add Member - %s - OCTO", app.AppName)
		}(),
		"User":        c.MustGet("User"),
		"appId":       appId,
		"app":         app,
		"userIdError": "",
	})
}

func (*UserAppService) AddConfirmEndpoint(c *gin.Context) {

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

	userIdParam := c.PostForm("userId")
	if userIdParam == "" {
		var app models.App
		if appId != 0 {
			if err := appDao.Get(&app, appId); err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		}

		c.HTML(http.StatusOK, "userAppAdd.tmpl", gin.H{
			"User":        c.MustGet("User"),
			"appId":       appId,
			"app":         app,
			"userIdError": "This feeld is required.",
		})
		return
	}

	roleTypeParam := c.PostForm("roleType")

	roleType, err := strconv.Atoi(roleTypeParam)
	if err != nil {
		c.String(http.StatusBadRequest, roleTypeParam+" is not roleTypeParam")
		return
	}
	if err := userAppDao.Add(appId, userIdParam, roleType); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%s/userapp", appIdParam))
}

func (*UserAppService) DeleteEndpoint(c *gin.Context) {
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

	userIdParam := c.PostForm("userId")
	if err := userAppDao.Delete(appId, userIdParam); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/a/%s/userapp", appIdParam))
}
