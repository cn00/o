package cache

import "testing"

func TestListKey(t *testing.T) {
	var listCache ListCache
	key := listCache.listKey(1, 2, 3, 4)
	if key != "list:1:2:3:4" {
		t.Fatal("key:", key)
	}
}
