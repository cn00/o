package cache

import (
	"fmt"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
)

type UrlCache struct{}

func (c *UrlCache) AssetBundleGet(appId int, versionId int, revisionId int, objectName string, maxRevisionId int) (octo.Url, bool) {
	key := c.assetBundleKey(appId, versionId, revisionId, objectName, maxRevisionId)
	return c.getUrl(key)
}

func (c *UrlCache) AssetBundleSet(appId int, versionId int, revisionId int, objectName string, maxRevisionId int, u octo.Url) {
	key := c.assetBundleKey(appId, versionId, revisionId, objectName, maxRevisionId)
	c.addUrl(key, u)
}

func (c *UrlCache) ResourceGet(appId int, versionId int, revisionId int, objectName string, maxRevisionId int) (octo.Url, bool) {
	key := c.resourceKey(appId, versionId, revisionId, objectName, maxRevisionId)
	return c.getUrl(key)
}

func (c *UrlCache) ResourceSet(appId int, versionId int, revisionId int, objectName string, maxRevisionId int, u octo.Url) {
	key := c.resourceKey(appId, versionId, revisionId, objectName, maxRevisionId)
	c.addUrl(key, u)
}

func (c *UrlCache) assetBundleKey(appId int, versionId int, revisionId int, objectName string, maxRevisionId int) string {
	return fmt.Sprintf("ab:%d:%d:%d:%s:%d", appId, versionId, revisionId, objectName, maxRevisionId)
}

func (c *UrlCache) resourceKey(appId int, versionId int, revisionId int, objectName string, maxRevisionId int) string {
	return fmt.Sprintf("r:%d:%d:%d:%s:%d", appId, versionId, revisionId, objectName, maxRevisionId)
}

func (c *UrlCache) getUrl(key string) (octo.Url, bool) {
	v, ok := urlCache.Get(key)
	if !ok {
		return octo.Url{}, false
	}
	return v.(octo.Url), true
}

func (c *UrlCache) addUrl(key string, u octo.Url) {
	urlCache.Add(key, u)
}
