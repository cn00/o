package controllers

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"octo/utils"
	"strconv"
	"strings"

	"octo-api/service"
	"octo/models"

	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var downloadService = &service.DownloadService{}

type ListEndpointDetail struct {
	Size      int    `json:"size"`
	Hash      uint32 `json:"hash"`
	EdgeCache bool   `json:"edgeCache"`
}

type DownloadListEndpointDetail struct {
	Count int `json:"count"`
}

func ListEndpoint(c *gin.Context) {
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

	// version中设置了密码密钥时，可以使用API不允许
	if version.ApiAesKey != "" {
		c.String(http.StatusBadRequest, "Must use encrypted API")
		return
	}

	database, err := downloadService.List(appId, v, r)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Etag", etag)
	if edgeCacheOK {
		enableEdgeCache(c)
	}

	c.Set("detail", ListEndpointDetail{
		Size:      len(database),
		Hash:      fnv32a(database),
		EdgeCache: edgeCacheOK,
	})
	writeProtobuf(c.Writer, http.StatusOK, database)
}

func ListAssetEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}

	maxRevision, err := downloadService.GetMaxRevisionInt(appId, v)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	header := c.Request.Header
	edgeCacheOK := checkAcceptHeader(header, appId)
	etag := listAssetEtag(appId, v, maxRevision)
	if edgeCacheOK && header.Get("If-None-Match") == etag {
		c.String(http.StatusNotModified, "")
		return
	}

	assetbundles, err := downloadService.ListAssetBundleWithAssets(appId, v)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Header("Etag", etag)
	if edgeCacheOK {
		enableEdgeCache(c)
	}

	c.JSON(http.StatusOK, struct {
		AssetBundles utils.List `json:"assetbundles"`
	}{assetbundles})
}

func ListTestEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}
	revision := c.Param("revision")
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	_, err = downloadService.List(appId, v, r)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "OK")
}

func FileDownloadEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}
	revision := c.Param("revision")
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	objectName := c.Param("objectName")

	res, bundleUrl, err := downloadService.GetAssetBundleUrl(appId, v, r, objectName)

	_, err = url.Parse(bundleUrl)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := errors.Cause(err).(*service.ObjectNotFoundError); ok {
		c.String(http.StatusNotFound, "")
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if c.Query("redirect") == "on" {
		c.Redirect(http.StatusSeeOther, bundleUrl)
	} else {
		writeProtobuf(c.Writer, http.StatusOK, res)
	}
}

func FileDownloadListEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}
	revision := c.Param("revision")
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	body := c.Request.Body
	x, err := ioutil.ReadAll(body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	s := string(x)
	objectNameList := splitObjectNameList(s)
	list, err := downloadService.GetAssetBundleUrlList(appId, v, r, objectNameList)
	if err != nil {
		log.Println("[INFO] objectNameList:", s)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Set("detail", DownloadListEndpointDetail{
		Count: len(objectNameList),
	})
	writeProtobuf(c.Writer, http.StatusOK, list)
}

func ResourceDownloadListEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}
	revision := c.Param("revision")
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	body := c.Request.Body
	x, err := ioutil.ReadAll(body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	s := string(x)
	objectNameList := splitObjectNameList(s)
	list, err := downloadService.GetResourceUrlList(appId, v, r, objectNameList)
	if err != nil {
		log.Println("[INFO] objectNameList:", s)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Set("detail", DownloadListEndpointDetail{
		Count: len(objectNameList),
	})
	writeProtobuf(c.Writer, http.StatusOK, list)
}

func ResourceDownloadEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}
	revision := c.Param("revision")
	log.Println(version, ":", revision)
	r, err := strconv.Atoi(revision)
	if err != nil {
		c.String(http.StatusBadRequest, revision+" is not revision")
		return
	}

	objectName := c.Param("objectName")

	res, url, err := downloadService.GetResourceUrl(appId, v, r, objectName)
	if _, ok := errors.Cause(err).(*service.ObjectNotFoundError); ok {
		c.String(http.StatusNotFound, "")
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else if c.Query("redirect") == "on" {
		c.Redirect(http.StatusSeeOther, url)
	} else {
		writeProtobuf(c.Writer, http.StatusOK, res)
	}
}

func MaxRevisionEndpoint(c *gin.Context) {
	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		c.String(http.StatusForbidden, "Invalid App")
		return
	}
	appId := app.AppId
	version := c.Param("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		c.String(http.StatusBadRequest, version+" is not version")
		return
	}

	res, err := downloadService.GetMaxRevision(appId, v)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	writeProtobuf(c.Writer, http.StatusOK, res)
}

func checkAcceptHeader(header http.Header, appId int) bool {
	accept := header.Get("Accept")
	ok := (accept == "application/x-protobuf,x-octo-app/"+strconv.Itoa(appId))
	if !ok {
		log.Println("[INFO] checkAcceptHeader: invalid:", accept)
	}
	return ok
}

func enableEdgeCache(c *gin.Context) {
	if gin.IsDebugging() {
		log.Println("[DEBUG] enable edge cache")
	}
	c.Header("Cache-Control", "public,max-age=0")
	c.Header("Vary", "Accept")
}

func writeProtobuf(w gin.ResponseWriter, statusCode int, body []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(statusCode)
	w.WriteHeaderNow()
	w.Flush()
	w.Write(body)
}

// Increment this version when list response is changed.
// If forget, the edge cache servers continue to serve the old response.
const listResponseVersion = 3
const listAssetResponseVersion = 1

func listEtag(appId, version, revision, maxRevision int) string {
	return fmt.Sprintf("\"%d.%d.%d.%d.%d\"", listResponseVersion, appId, version, revision, maxRevision)
}

func listAssetEtag(appId, version, maxRevision int) string {
	return fmt.Sprintf("\"la%d.%d.%d.%d.%d\"", listAssetResponseVersion, appId, version, maxRevision)
}

func splitObjectNameList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func fnv32a(b []byte) uint32 {
	h := fnv.New32a()
	h.Write(b)
	return h.Sum32()
}
