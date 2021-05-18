package service

import "testing"

func TestDownloadServiceCreateCachedUrl(t *testing.T) {
	var downloadService DownloadService
	const url = "https://www.googleapis.com/download/storage/v1/b/golangtest--1-assetbundle/o/VAKRY0?generation=1465888064614000&alt=media"
	const cdnUrl = "https://storage.googleapis.com"
	cachedUrl, err := downloadService.createUrl(url, cdnUrl)
	if err != nil {
		t.Fatal(err)
	}
	if cachedUrl != "https://storage.googleapis.com/golangtest--1-assetbundle/VAKRY0?generation=1465888064614000&alt=media" {
		t.Fatal("cachedUrl:", cachedUrl)
	}
}

func BenchmarkDownloadServiceCreateCachedUrl(b *testing.B) {
	var downloadService DownloadService
	const url = "https://www.googleapis.com/download/storage/v1/b/golangtest--1-assetbundle/o/VAKRY0?generation=1465888064614000&alt=media"
	const cdnUrl = "https://storage.googleapis.com"
	for i := 0; i < b.N; i++ {
		_, err := downloadService.createUrl(url, cdnUrl)
		if err != nil {
			b.Fatal(err)
		}
	}
}
