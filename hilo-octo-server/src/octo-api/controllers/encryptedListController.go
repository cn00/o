package controllers

import (
	"crypto/sha256"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"octo/models"
	"octo/utils"
	"os"
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

	//// TODO:
	//// version中未设置密码密钥时，将加密API不能使用
	//if version.ApiAesKey == "" {
	//	c.String(http.StatusBadRequest, "ApiAesKey is null")
	//}

	database, err := downloadService.List(appId, v, r)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ioutil.WriteFile("database.protobuf.bin", database, os.ModePerm)

	key := sha256.Sum256([]byte(version.ApiAesKey))
	cipherDatabase, err := utils.EncryptAes256((database), key)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	cipherDatabase = database // TODO: 先不加密
	
	ioutil.WriteFile("database.protobuf.aes256.bin", cipherDatabase, os.ModePerm)

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