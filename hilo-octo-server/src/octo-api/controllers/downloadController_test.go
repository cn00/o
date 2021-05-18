package controllers

import "testing"

func TestListEtag(t *testing.T) {
	etag := listEtag(1, 2, 3, 4)
	if etag != "\"3.1.2.3.4\"" {
		t.Fatal("etag:", etag)
	}
}
