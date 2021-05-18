package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"

	"net/url"
)

func GetUploadVersionId(pathURL *url.URL, originalBucketName string) (int, error) {
	// Parse uploadVersionId
	paths := strings.Split(pathURL.Path, "/")

	// get bucket name from url
	var bucketName string
	for i := range paths {
		path := paths[i]
		if strings.HasPrefix(path, originalBucketName) {
			bucketName = path
			break
		}
	}

	if len(bucketName) == 0 {
		return 0, errors.New("no have bucket name " + pathURL.Path)
	}

	if !(strings.Contains(bucketName, "assetbundle") || strings.Contains(bucketName, "resources")) {
		return 0, errors.New("no have a suffix that type(assetbundle, resources) on file  of " + pathURL.Path)
	}

	// get version part from bucket name
	bucketName = strings.TrimPrefix(bucketName, originalBucketName)
	// get version with cutting hyphen
	bucketArrays := strings.Split(bucketName, "-")

	if len(bucketArrays) <= 1 {
		return 0, errors.New("no have version id on bucket name")
	}
	// process get version id
	uploadVersionIdInt, err := strconv.Atoi(bucketArrays[1])
	if err != nil {
		return 0, fmt.Errorf("BucketsTable buckename : %s, file url %s, error:%v", originalBucketName, pathURL, err)
	}

	return uploadVersionIdInt, nil

}
