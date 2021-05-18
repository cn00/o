package cache

import "fmt"

type ListCache struct{}

func (c *ListCache) ListGet(appId int, versionId int, revisionId int, maxRevisionId int) ([]byte, bool) {
	key := c.listKey(appId, versionId, revisionId, maxRevisionId)
	return c.getByteSlice(key)
}

func (c *ListCache) ListSet(appId int, versionId int, revisionId int, maxRevisionId int, b []byte) {
	key := c.listKey(appId, versionId, revisionId, maxRevisionId)
	listCache.Add(key, b)
}

func (c *ListCache) listKey(appId int, versionId int, revisionId int, maxRevisionId int) string {
	return fmt.Sprintf("list:%d:%d:%d:%d", appId, versionId, revisionId, maxRevisionId)
}

func (c *ListCache) getByteSlice(key string) ([]byte, bool) {
	v, ok := listCache.Get(key)
	if !ok {
		return nil, false
	}
	return v.([]byte), true
}
