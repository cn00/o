package controllers

import (
	"crypto/sha256"
	"github.com/gin-gonic/gin"
	"net/http"
	"octo/models"
	"octo/utils"
	"strconv"
)

func EncryptedListEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	versionParam := c.Param("version")
	v, err := strconv.Atoi(versionParam)
	if err != nil {
		c.String(http.StatusBadRequest, versionParam+" is not version")
		return
	}
	revision := c.Param("revision")
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	version, err := downloadService.GetVersion(appId, v)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	header := c.Request.Header
	edgeCacheOK := checkAcceptHeader(header, appId)
	etag := listEtag(appId, v, r, version.MaxRevision)
	if edgeCacheOK && header.Get("If-None-Match") == etag {
		c.String(http.StatusNotModified, "")
		return
	}

	// versionに暗号鍵が設定されていない場合、暗号化APIは使えない
	if version.ApiAesKey == "" {
		c.String(http.StatusBadRequest, "Doesn't support encrypt")
	}

	database, err := downloadService.List(appId, v, r)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	key := sha256.Sum256([]byte(version.ApiAesKey))
	cipherDatabase, err := utils.EncryptAes256(database, key)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Etag", etag)
	if edgeCacheOK {
		enableEdgeCache(c)
	}

	c.Set("detail", ListEndpointDetail{
		Size:      len(cipherDatabase),
		Hash:      fnv32a(cipherDatabase),
		EdgeCache: edgeCacheOK,
	})
	writeProtobuf(c.Writer, http.StatusOK, cipherDatabase)
}