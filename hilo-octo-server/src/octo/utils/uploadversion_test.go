package utils

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestUploadVersion_SuccessGetUploadVersionId(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-101-assetbundle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 101, uploadVersionId)

	parseURL, err = url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-1-101-assetbundle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName = "melo-stg-1"

	uploadVersionId, err = GetUploadVersionId(parseURL, orgBucketName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 101, uploadVersionId)

}

func TestUploadVersion_SuccessGetUploadVersionIdResources(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-101-resources/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 101, uploadVersionId)

}

func TestUploadVersion_SuccessGetUploadVersionIdAnotherURL(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/melo-stg-101-resources/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 101, uploadVersionId)

}

func TestUploadVersion_FailGetUploadVersionIdResources(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-101-rep/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)

	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)

}

func TestUploadVersion_FailGetUploadVersionId(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-101-asset/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)

}

func TestUploadVersion_FailPathGetUploadVersionId(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/v1/b/-101-assetbundle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)
}

func TestGetUploadVersion_FailBucketNameNotMatching(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/v1/b/melo-stg-101-assetbunle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "nazca"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)
}

func TestGetUploadVersion_FailGetVersionIdOnBukcet(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-aaa-assetbunle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)
}

func TestGetUploadVersion_FailParseBukcetName(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-aaa-aaa-assetbunle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)
}

func TestGetUploadVersion_FailParseBukcetName2(t *testing.T) {
	parseURL, err := url.Parse("https://www.googleapis.com/download/storage/v1/b/melo-stg-assetbunle/o/00BVwD?generation=1462175496389000&alt=media")
	if err != nil {
		t.Fatal(err)
	}
	orgBucketName := "melo-stg"

	uploadVersionId, err := GetUploadVersionId(parseURL, orgBucketName)
	assert.Equal(t, 0, uploadVersionId)
	assert.NotNil(t, err)
}
