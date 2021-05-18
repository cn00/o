package controllers

import (
	"net/http"

	"octo/controllers"
	"octo/utils"

	"github.com/gin-gonic/gin"
	"octo-api/config"
)

var authMiddleware = &AuthMiddleware{}
var conf = config.LoadConfig()

func InitRooter(e *gin.Engine) {
	r := &utils.RouterGroup{
		RouterGroup: *e.Group("", gin.ErrorLogger()),
	}
	r.GETorHEAD("/status", controllers.StatusEndpoint)
	r.GETorHEAD("/status/db", controllers.StatusDBEndpoint)
	r.GETorHEAD("/status/read_only", StatusReadOnlyEndpoint)
	r.Static("/static", "./static")

	initV1PublicRoute(e)
	initV1AdminRoute(e)
	initV2PublicRoute(e)
	initV2AdminRoute(e)
}

func initV1PublicRoute(e *gin.Engine) {
	r := &utils.RouterGroup{
		RouterGroup: *e.Group("/v1",
			gin.ErrorLoggerT(gin.ErrorTypePublic),
			authMiddleware.ClientAuthV1,
		),
	}
	r.GETorHEAD("list/:version/:revision", ListEndpoint)
	r.GETorHEAD("listasset/:version", ListAssetEndpoint)
	r.GETorHEAD("ab/:version/:revision/:objectName", FileDownloadEndpoint)
	r.POST("ab/:version/:revision", FileDownloadListEndpoint)
	r.GETorHEAD("r/:version/:revision/:objectName", ResourceDownloadEndpoint)
	r.POST("r/:version/:revision", ResourceDownloadListEndpoint)
	r.GETorHEAD("revision/:version", MaxRevisionEndpoint)
	r.POST("er", ErrorReportEndpoint)
}

func initV1AdminRoute(e *gin.Engine) {
	r := &utils.RouterGroup{
		RouterGroup: *e.Group("/v1",
			cliErrorLogger,
			authMiddleware.AppAuthV1,
			checkCliVersion,
		),
	}

	r.GETorHEAD("check/diff/ab/:versionId/:revisionId/:targetRevisionId", CheckDiffAssetBundleEndpoint)
	r.GETorHEAD("check/diff/r/:versionId/:revisionId/:targetRevisionId", CheckDiffResourceEndpoint)

	wr := &utils.RouterGroup{
		RouterGroup: *r.Group("", checkReadOnly),
	}

	wr.GETorHEAD("upload/list/:version", UploadListEndpoint)
	wr.GETorHEAD("upload/start", UploadStartEndpoint)
	wr.POST("upload/all/:version", UploadAllEndpoint)
	wr.POST("upload/all/:version/notag", UploadAllNoTagEndpoint)

	wr.GETorHEAD("resource/upload/list/:version", ResourceUploadListEndpoint)
	wr.GETorHEAD("resource/upload/start", ResourceUploadStartEndpoint)
	wr.POST("resource/upload/all/:version", ResourceUploadAllEndpoint)
	wr.POST("resource/upload/all/:version/notag", ResourceUploadAllNoTagEndpoint)

	wr.POST("tag/ab/:version", TagAssetBundleEndpoint)
	wr.POST("tag/r/:version", TagResourceEndpoint)

	wr.POST("remove/tag/ab/:version", RemoveTagAssetBundleEndpoint)
	wr.POST("remove/tag/r/:version", RemoveTagResourceEndpoint)

	wr.POST("delete/ab/:version", DeleteAssetBundleEndpoint)
	wr.POST("delete/r/:version", DeleteResourceEndpoint)

	wr.POST("sync/:dstVersionId/:revisionId/diff/:srcAppId/:srcVersionId", DiffSyncEndpoint)
}

func cliErrorLogger(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		c.JSON(-1, struct {
			Error string `json:"Error"`
		}{
			Error: c.Errors.String(),
		})
	}
}

func initV2PublicRoute(e *gin.Engine) {
	r := &utils.RouterGroup{
		RouterGroup: *e.Group("/v2/pub/a/:appId",
			gin.ErrorLoggerT(gin.ErrorTypePublic),
			authMiddleware.ClientAuthV2,
		),
	}
	r.GETorHEAD("/v/:version/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GETorHEAD("/v/:version/list/:revision", EncryptedListEndpoint)
}

func initV2AdminRoute(e *gin.Engine) {
	r := &utils.RouterGroup{
		RouterGroup: *e.Group("/v2/admin/a/:appId",
			gin.ErrorLogger(),
			authMiddleware.AppAuthV2,
			checkCliVersion,
		),
	}
	r.GETorHEAD("list/ab/:versionId", CheckListAssetBundleEndpoint)
	r.GETorHEAD("list/r/:versionId", CheckListResourceEndpoint)
	r.GETorHEAD("list/ab/:versionId/:fromDate/:toDate", CheckListRangeAssetBundleEndpoint)
	r.POST("url", CheckURLEndpoint)
	r.POST("sync/:dstAppId/:dstVersionId/:revisionId/diff/:srcAppId/:srcVersionId", DiffSyncEndpointV2)

	wr := &utils.RouterGroup{
		RouterGroup: *r.Group("", checkReadOnly),
	}
	wr.POST("copy/:type", CopyEndpoint)
}
