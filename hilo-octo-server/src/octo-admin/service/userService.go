package service

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"

	"octo/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/wataru420/contrib/sessions"
)

var userDao = &models.UserDao{}

var userService = &UserService{}

type UserService struct {
}

func (*UserService) LoginEndpoint(c *gin.Context) {
	id := c.PostForm("id")
	password := c.PostForm("password")

	user, err := userDao.Get(id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Println("not ldap")
	h := sha256.Sum256([]byte(password))
	sha256 := base64.StdEncoding.EncodeToString(h[:])
	if user.Password != sha256 {
		c.AbortWithError(http.StatusInternalServerError, errors.New("wrong id or password."))
		return
	}

	session := sessions.Default(c)
	session.Set("id", id)
	v := session.Get("fromUrl")
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if v != nil {
		session.Delete("fromUrl")
		c.Redirect(http.StatusSeeOther, v.(string))
	} else {
		c.Redirect(http.StatusSeeOther, "/")
	}
}

func (*UserService) CheckLogin(c *gin.Context) {
	session := sessions.Default(c)
	v := session.Get("id")
	if v == nil {
		log.Println("log outed")
		var fromUrl string
		fromUrl = c.Request.URL.String()
		session.Set("fromUrl", fromUrl)
		if err := session.Save(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			c.Abort()
			return
		}
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
	} else {
		user, err := userDao.Get(v.(string))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			c.Abort()
			return
		}
		c.Set("User", user)
		userApps, err := userAppDao.GetByUserId(v.(string))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			c.Abort()
			return
		}
		c.Set("UserApps", userApps)
	}
}

func (s *UserService) CheckAuthority(c *gin.Context, appId int, roleType models.UserAppRoleType) bool {
	userApps := c.MustGet("UserApps").(models.UserApps)
	return s.checkAuthority(userApps, appId, roleType)
}

func (*UserService) checkAuthority(userApps models.UserApps, appId int, roleType models.UserAppRoleType) bool {
	for _, userApp := range userApps {
		if userApp.AppId == 0 || userApp.AppId == appId {
			switch models.UserAppRoleType(userApp.RoleType) {
			case models.UserRoleTypeAdmin:
				if roleType == models.UserRoleTypeAdmin {
					return true
				}
				fallthrough
			case models.UserRoleTypeUser:
				if roleType == models.UserRoleTypeUser {
					return true
				}
				fallthrough
			case models.UserRoleTypeReader:
				if roleType == models.UserRoleTypeReader {
					return true
				}
				fallthrough
			default:
				log.Printf("need RoleType %d but user is %d\n", int(roleType), userApp.RoleType)
			}
		}
	}
	return false
}

func (*UserService) LogoutEndpoint(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}
