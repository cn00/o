package controllers

import (
	"math"
	"net/http"
	"strconv"

	"octo-api/service"
	"octo/models"

	"octo/utils"

	"github.com/gin-gonic/gin"
)

var checkService = &service.CheckService{}

func CheckListAssetBundleEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)
	version, err := getParamVersionId(c)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}

	list, err := checkService.ListActiveAssetBundle(app.AppId, version)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func CheckListRangeAssetBundleEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)
	version, err := getParamVersionId(c)

	fromDate := c.Param("fromDate")
	toDate := c.Param("toDate")

	err, _ = utils.CheckFromToDateFormat(fromDate, toDate)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}

	list, err := checkService.ListRangeActiveAssetBundle(app.AppId, version, fromDate, toDate)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, list)

}

func CheckListResourceEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)
	version, err := getParamVersionId(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}

	list, err := checkService.ListActiveResource(app.AppId, version)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func CheckDiffAssetBundleEndpoint(c *gin.Context) {

	var res struct {
		List  interface{}
		Error string
	}

	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		res.Error = "Invalid App"
		c.JSON(http.StatusForbidden, res)
		return
	}

	versionIdParam := c.Param("versionId")
	version, err := strconv.Atoi(versionIdParam)
	if err != nil {
		res.Error = versionIdParam + " is not version"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	revisionIdParam := c.Param("revisionId")
	revision, err := strconv.Atoi(revisionIdParam)
	if err != nil {
		res.Error = revisionIdParam + " is not revision"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	targetRevisionIdParam := c.Param("targetRevisionId")
	trevision, err := strconv.Atoi(targetRevisionIdParam)
	if err != nil {
		res.Error = targetRevisionIdParam + " is not revision"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	list, err := checkService.DiffAssetBundle(app.AppId, version, revision, trevision)
	if err != nil {
		c.Error(err)
		res.Error = err.Error()
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res.List = list
	c.JSON(http.StatusOK, res)
}

func CheckDiffResourceEndpoint(c *gin.Context) {

	var res struct {
		List  interface{}
		Error string
	}

	val, _ := c.Get("app")
	app, ok := val.(models.App)
	if !ok {
		res.Error = "Invalid App"
		c.JSON(http.StatusForbidden, res)
		return
	}

	versionIdParam := c.Param("versionId")
	version, err := strconv.Atoi(versionIdParam)
	if err != nil {
		res.Error = versionIdParam + " is not version"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	revisionIdParam := c.Param("revisionId")
	revision, err := strconv.Atoi(revisionIdParam)
	if err != nil {
		res.Error = revisionIdParam + " is not revision"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	targetRevisionIdParam := c.Param("targetRevisionId")
	trevision, err := strconv.Atoi(targetRevisionIdParam)
	if err != nil {
		res.Error = targetRevisionIdParam + " is not revision"
		c.JSON(http.StatusBadRequest, res)
		return
	}

	list, err := checkService.DiffResource(app.AppId, version, revision, trevision)
	if err != nil {
		c.Error(err)
		res.Error = err.Error()
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res.List = list
	c.JSON(http.StatusOK, res)
}

func CheckURLEndpoint(c *gin.Context) {
	app := c.MustGet("app").(models.App)

	var json struct {
		VersionID    int      `json:"version_id"`
		Revision     int      `json:"revision"`
		AssetBundles []string `json:"assetbundles"`
		Resources    []string `json:"resources"`
	}
	if err := c.BindJSON(&json); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypePublic)
		return
	}
	if json.Revision == 0 {
		json.Revision = math.MaxInt32
	}

	var res struct {
		AssetBundles map[string]*string `json:"assetbundles"`
		Resources    map[string]*string `json:"resources"`
	}

	if len(json.AssetBundles) > 0 {
		assetBundles, err := downloadService.GetAssetBundleUrlListByName(app.AppId, json.VersionID, json.Revision, json.AssetBundles)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err).SetType(gin.ErrorTypePublic)
			return
		}
		res.AssetBundles = make(map[string]*string, len(json.AssetBundles))
		for i, a := range json.AssetBundles {
			if assetBundles[i] != nil {
				res.AssetBundles[a] = assetBundles[i].Url
			} else {
				res.AssetBundles[a] = nil
			}
		}
	}

	if len(json.Resources) > 0 {
		resources, err := downloadService.GetResourceUrlListByName(app.AppId, json.VersionID, json.Revision, json.Resources)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err).SetType(gin.ErrorTypePublic)
			return
		}
		res.Resources = make(map[string]*string, len(json.Resources))
		for i, r := range json.Resources{

			if resources[i] != nil {
				res.Resources[r] = resources[i].Url
			} else {
				res.Resources[r] = nil
			}
		}
	}
	c.JSON(http.StatusOK, res)
	return
}
