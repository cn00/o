package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"octo/models"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/wataru420/contrib/sessions"
	"golang.org/x/oauth2"
)

type OauthGoogleService struct {
}

var oauthGoogleConfig *oauth2.Config

func OauthGoogleSetup(config *oauth2.Config) {
	oauthGoogleConfig = config
}

func IsOautGoogle() bool {
	return oauthGoogleConfig.ClientID != ""
}

func (*OauthGoogleService) LoginEndpoint(c *gin.Context) {
	url := oauthGoogleConfig.AuthCodeURL("state")
	c.Redirect(http.StatusSeeOther, url)
}

func (*OauthGoogleService) OauthEndpoint(c *gin.Context) {
	//stateの値のチェック
	state := c.Query("state")
	log.Println(state)

	//errorチェック
	apiError := c.Query("error")
	if apiError != "" {
		c.AbortWithError(http.StatusInternalServerError, errors.New(apiError))
		return
	}

	//アクセストークン取得
	authcode := c.Query("code")

	tok, err := oauthGoogleConfig.Exchange(oauth2.NoContext, authcode)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap(err, "exchange error"))
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + tok.AccessToken)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap(err, "get userinfo error"))
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.Wrap(err, "read all error"))
		return
	}

	var f interface{}
	json.Unmarshal(contents, &f)
	m := f.(map[string]interface{})

	//ユーザのプロフィール情報を取得
	email := m["email"].(string)

	//ユーザ登録等
	user, err := userDao.Get(email)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if (user == models.User{}) {
		user.Email = email
		user.UserId = email
		user.AuthType = int(models.UserAuthTypeOauthGoogle)
		if err := userDao.Insert(&user); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	//sessionに格納してリダイレクト
	session := sessions.Default(c)
	session.Set("id", email)
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
