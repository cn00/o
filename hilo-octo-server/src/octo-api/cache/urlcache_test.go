package cache

import "testing"

func TestAssetBundleKey(t *testing.T) {
	var urlCache UrlCache
	key := urlCache.assetBundleKey(1, 2, 3, "objectName", 4)
	if key != "ab:1:2:3:objectName:4" {
		t.Fatal("key:", key)
	}
}

func TestResourceKey(t *testing.T) {
	var urlCache UrlCache
	key := urlCache.resourceKey(1, 2, 3, "objectName", 4)
	if key != "r:1:2:3:objectName:4" {
		t.Fatal("fail:", key)
	}
}
