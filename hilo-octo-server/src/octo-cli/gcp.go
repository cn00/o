/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Binary storage-sample creates a new bucket, performs all of its operations
// within that bucket, and then cleans up after itself if nothing fails along the way.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

// GoogleCloudStorage : A Google Cloud Storage Information
type GoogleCloudStorage struct {
	ProjectID  string
	BucketName string
	Location   string
}

type GCPCORSValues struct {
	MaxAge          time.Duration `json:"maxAge"`
	Methods         []string      `json:"methods"`
	Origins         []string      `json:"origins"`
	ResponseHeaders []string      `json:"responseHeaders"`
}

func NewGoogleCloud() *GoogleCloudStorage {
	return NewGoogleCloudStorage(Conf.Gcs.ProjectID, Conf.Gcs.BucketName, Conf.Gcs.Location)
}

// NewGoogleCloudStorage creates a New GCS
func NewGoogleCloudStorage(projectID string, bucketName string, location string) *GoogleCloudStorage {
	log.Printf("NewGCS projectId : %v, bucketName : %v, location : %v\n", projectID, bucketName, location)
	gcs := &GoogleCloudStorage{projectID, bucketName, location}
	return gcs
}
func (gcs *GoogleCloudStorage) startUploadToOSS(priority int, tags cli.StringSlice, buildNumber, uploadType string, fileMap *FileMap) int {
	return startUploadToGCP(gcs, priority, tags, buildNumber, uploadType, fileMap)
}
func (gcs *GoogleCloudStorage) createBucket() {

	ctx := context.Background()

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(gcs.BucketName)
	attrs, err := bucket.Attrs(ctx)

	if err != nil {
		// Creates the new bucket.
		// OCTOではVersioningを使用しているため、必ずtrueを設定すること
		if err := bucket.Create(ctx, gcs.ProjectID, &storage.BucketAttrs{
			Location:          "asia",
			DefaultObjectACL:  []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}},
			VersioningEnabled: true,
		}); err != nil {
			log.Fatalf("Failed to create bucket for projectId : %v. Error : %v \n", gcs.ProjectID, err)
			return
		}

		log.Printf("Bucket %v created.\n", gcs.BucketName)
	} else {
		log.Printf("Already has bucket %s\n", attrs.Name)
	}
}

func (gcs *GoogleCloudStorage) setCORSWithJSON(jsonStr string) {
	corsV := GCPCORSValues{}
	err := json.Unmarshal([]byte(jsonStr), &corsV)
	if err != nil {
		log.Fatalf("Failed to unmarshal cors json : %v", err)
		return
	}
	gcs.setCORS(corsV)
}

func (gcs *GoogleCloudStorage) setCORS(corsV GCPCORSValues) {

	cors := []storage.CORS{{
		MaxAge:          corsV.MaxAge,
		Methods:         corsV.Methods,
		Origins:         corsV.Origins,
		ResponseHeaders: corsV.ResponseHeaders,
	}}
	ctx := context.Background()

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(gcs.BucketName)
	_, err = bucket.Attrs(ctx)

	if err != nil {
		log.Fatalf("Not FoundBucket: %v", err)
	}

	_, err = bucket.Update(ctx, storage.BucketAttrsToUpdate{CORS: cors})
	if err != nil {
		log.Fatalf("Set CROS Error:  %v", err)
	}

	log.Printf("Successful Update CORS :%v", corsV)

}

func (gcs *GoogleCloudStorage) upload(fileName string, objectName string) (string, error) {
	ctx := context.Background()

	f, err := os.Open(fileName)

	if err != nil {
		return "", err
	}
	defer f.Close()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	obj := client.Bucket(gcs.BucketName).Object(objectName)

	wc := obj.NewWriter(ctx)

	if _, err = io.Copy(wc, f); err != nil {
		return "", err
	}

	if err := wc.Close(); err != nil {
		return "", err
	}

	objAttr, err := obj.Attrs(ctx)
	if err != nil {
		return "", err
	}
	mediaLink := objAttr.MediaLink
	fmt.Printf("Created object %v at location %v\n\n", objAttr.Name, mediaLink)

	// update acl
	acl := obj.ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		log.Printf("Failed to set ACL: %v\n", err)
		return "", err
	}
	return mediaLink, nil
}

func (gcs *GoogleCloudStorage) uploadWithChan(filePath string, fileName string, assetBundleName string, objectName string, dependencies []string) GCSFile {
	do := func() GCSFile {
		ctx := context.Background()
		f, err := os.Open(filePath)

		defer f.Close()
		if err != nil {
			log.Println("gcs", filePath, err)
			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}
			return file

		}

		client, err := storage.NewClient(ctx)

		if err != nil {
			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}
			return file
		}
		obj := client.Bucket(gcs.BucketName).Object(objectName)

		wc := obj.NewWriter(ctx)

		if _, err = io.Copy(wc, f); err != nil {
			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}
			return file
		}

		if err := wc.Close(); err != nil {

			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}

			return file
		}

		objAttr, err := obj.Attrs(ctx)
		if err != nil {
			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}

			return file
		}
		mediaLink := objAttr.MediaLink
		fmt.Printf("%v Created object %v at location %v\n", fileName, objAttr.Name, mediaLink)

		// update acl
		acl := obj.ACL()
		if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			log.Printf("Failed to set ACL: %v\n", err)
			file := GCSFile{filePath, fileName, assetBundleName, objectName, "", dependencies, err}

			return file
		}

		file := GCSFile{filePath, fileName, assetBundleName, objectName, mediaLink, dependencies, nil}

		return file
	}
	return runWithRetry(do)
}

func runWithRetry(call func() GCSFile) GCSFile {
	var file GCSFile
	for i := 0; i < 3; i++ {
		file = call()
		if file.Err != nil {
			if strings.Contains(file.Err.Error(), "unexpected EOF") {
				continue
			}
			if strings.Contains(file.Err.Error(), "i/o timeout") {
				continue
			}
		}
		return file
	}
	return file
}
