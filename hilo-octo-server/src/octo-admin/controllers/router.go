package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"html/template"
	"log"
	"net/http"

	"octo-admin/config"
	"octo-admin/service"
	"octo/controllers"
	"octo/utils"

	"hilo-octo-proto/go/octo"
	"github.com/gin-gonic/gin"
	"github.com/wataru420/contrib/sessions"
)

var versionService = &service.VersionService{}
var userAppService = &service.UserAppService{}
var appService = &service.AppService{}
var oauthGoogleService = &service.OauthGoogleService{}
var bucketService = &service.BucketService{}
var gcsService = &service.GcsService{}
var conf = config.LoadConfig()

func InitRooter(e *gin.Engine) {

	funcMap := template.FuncMap{
		"add":               func(a, b int) int { return a + b },
		"sub":               func(a, b int) int { return a - b },
		"mul":               func(a, b int) int { return a * b },
		"div":               func(a, b int) int { return a / b },
		"mod":               func(a, b int) int { return a % b },
		"md5":               (func(string) string)(GetMD5Hash),
		"splitTags":         (func(string) []string)(utils.SplitTags),
		"splitDependencies": (func(string) ([]int, error))(utils.SplitDependencies),
		"dataState":         func(a int) string { return octo.Data_State(a).String() },
	}
	tmpl := template.Must(template.New("projectViews").Funcs(funcMap).ParseGlob("templates/admin/*.tmpl"))
	e.SetHTMLTemplate(tmpl)

	store := sessions.NewCookieStore(getCookieStoreHashKey())
	store.Options(sessions.Options{MaxAge: 60 * 60 * 24, Path: "/"})
	r := &utils.RouterGroup{RouterGroup: *e.Group("", sessions.Sessions("mysession", store))}
	checkLoginR := &utils.RouterGroup{RouterGroup: *r.Group("", userService.CheckLogin)}

	r.GETorHEAD("/status", controllers.StatusEndpoint)
	r.GETorHEAD("/status/db", controllers.StatusDBEndpoint)

	r.StaticFile("/favicon.ico", "./static/favicons/favicon.ico")
	r.StaticFile("/apple-touch-icon-57x57.png", "./static/favicons/apple-touch-icon-57x57.png")
	r.StaticFile("/apple-touch-icon-60x60.png", "./static/favicons/apple-touch-icon-60x60.png")
	r.StaticFile("/apple-touch-icon-72x72.png", "./static/favicons/apple-touch-icon-72x72.png")
	r.StaticFile("/apple-touch-icon-76x76.png", "./static/favicons/apple-touch-icon-76x76.png")
	r.StaticFile("/apple-touch-icon-114x114.png", "./static/favicons/apple-touch-icon-114x114.png")
	r.StaticFile("/apple-touch-icon-120x120.png", "./static/favicons/apple-touch-icon-120x120.png")
	r.StaticFile("/apple-touch-icon-144x144.png", "./static/favicons/apple-touch-icon-144x144.png")
	r.StaticFile("/apple-touch-icon-152x152.png", "./static/favicons/apple-touch-icon-152x152.png")
	r.StaticFile("/apple-touch-icon-180x180.png", "./static/favicons/apple-touch-icon-180x180.png")
	r.StaticFile("/android-chrome-192x192.png", "./static/favicons/android-chrome-192x192.png")
	r.StaticFile("/favicon-48x48.png", "./static/favicons/favicon-48x48.png")
	r.StaticFile("/favicon-96x96.png", "./static/favicons/favicon-96x96.png")
	r.StaticFile("/favicon-16x16.png", "./static/favicons/favicon-16x16.png")
	r.StaticFile("/favicon-32x32.png", "./static/favicons/favicon-32x32.png")
	r.StaticFile("/manifest.json", "./static/favicons/manifest.json")
	r.StaticFile("/mstile-144x144.png", "./static/favicons/mstile-144x144.png")

	r.Static("/static", "./static")

	checkLoginR.GETorHEAD("/", appMainEndpoint)

	r.GETorHEAD("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"isOauthGoogle": service.IsOautGoogle(),
		})
	})
	r.POST("/login", userService.LoginEndpoint)
	checkLoginR.GETorHEAD("/logout", userService.LogoutEndpoint)

	r.POST("/google/login", oauthGoogleService.LoginEndpoint)
	r.GETorHEAD("/google/oauth", oauthGoogleService.OauthEndpoint)

	checkLoginWriteR := &utils.RouterGroup{RouterGroup: *checkLoginR.Group("", checkReadOnly)}

	checkLoginR.GETorHEAD("/a/:appid/detail", appDetailEndpoint)
	checkLoginWriteR.POST("/a/:appid/update", appUpdateEndpoint)
	checkLoginWriteR.POST("/a/:appid/delete", appDeleteEncpoint)
	checkLoginR.GETorHEAD("/newApp", appNewEncpoint)
	checkLoginWriteR.POST("/insertApp", appInsertEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/v/:version/detail", versionDetailEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/update", versionUpdateEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/delete", versionDeleteEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/userapp", userAppService.ListEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/userapp/add", userAppService.AddEndpoint)
	checkLoginWriteR.POST("/a/:appid/userapp/add", userAppService.AddConfirmEndpoint)
	checkLoginWriteR.POST("/a/:appid/userapp/delete", userAppService.DeleteEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/v/:version/file/:fileid/detail", assetBundleDetailEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/file/:fileid/update", assetBundleUpdateEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/file/:fileid/delete", assetBundleDeleteEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/file", assetBundleListEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/fileDiff", assetBundleDiffEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/deleteSelectedFile", assetBundleDeleteSelectedFileEndpoint)
	checkLoginWriteR.POST("/a/:appid/copySelectedFile", assetBundleCopySelectedFileEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/fileCsvOutput", assetBundleCsvOutputEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/fileDiffCsvOutput", assetBundleDiffCsvOutputEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/v/:version/resource/:fileid/detail", resourceDetailEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/resource/:fileid/update", resourceUpdateEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/resource/:fileid/delete", resourceDeleteEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/resource", resourceListEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/resourceDiff", resourceDiffEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/resourceDiff/:fileid", resourceDiffFileEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/deleteSelectedResource", resourceDeleteSelectedFileEndpoint)
	checkLoginWriteR.POST("/a/:appid/v/:version/copySelectedResource", resourceCopySelectedFileEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/resourceCsvOutput", resourceCsvOutputEndpoint)
	checkLoginR.GETorHEAD("/a/:appid/v/:version/resourceDiffCsvOutput", resourceDiffCsvOutputEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/v/:version/maintenance/ta/:tappid/tv/:tversion", makeDiffSqlEndpoint)

	checkLoginR.GETorHEAD("/a/:appid/env", envListEndPoint)
	checkLoginR.GETorHEAD("/a/:appid/env/add", envAddEndPoint)
	checkLoginWriteR.POST("/a/:appid/env/add", envCreateEndPoint)
	//checkLoginWriteR.POST("/a/:appid/env/delete", envDeleteEndPoint)
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCookieStoreHashKey() []byte {
	conf := config.LoadConfig()
	secret := conf.Admin.CookieSecret
	log.Println("Cookie Secret:", secret)
	key, err := hex.DecodeString(secret)
	if err != nil {
		panic(err)
	}
	return key
}

func checkReadOnly(c *gin.Context) {
	conf := config.LoadConfig()
	if conf.Admin.ReadOnly {
		c.AbortWithError(http.StatusServiceUnavailable,
			errors.New("Admin is read only")).SetType(gin.ErrorTypePublic)
		return
	}
}
