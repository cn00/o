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
package service

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
	storage "google.golang.org/api/storage/v1"
)

const (
	// This scope allows the application full control over resources in Google Cloud Storage
	scope = storage.DevstorageFullControlScope
)

type GoogleCloudStorage struct {
	ProjectId  string
	BucketName string
	Location   string
}

func NewGoogleCloudStorage(projectId string, bucketName string, location string) *GoogleCloudStorage {
	gcs := &GoogleCloudStorage{projectId, bucketName, location}
	return gcs
}

func (self *GoogleCloudStorage) fatalf(service *storage.Service, errorMessage string, args ...interface{}) {
	log.Printf("Dying with error:\n"+errorMessage, args...)
}

func (self *GoogleCloudStorage) createBucket() {
	client, err := google.DefaultClient(context.Background(), scope, compute.ComputeScope)
	if err != nil {
		log.Printf("Unable to get default client: %v", err)
	}
	service, err := storage.New(client)
	if err != nil {
		log.Printf("Unable to create storage service: %v", err)
	}

	// If the bucket already exists and the user has access, warn the user, but don't try to create it.
	if _, err := service.Buckets.Get(self.BucketName).Do(); err == nil {
		fmt.Printf("Bucket %s already exists - skipping buckets.insert call.\n\n", self.BucketName)
	} else {
		fmt.Printf("Failed get bucket %s: %v\n", self.BucketName, err)
		if res, err := service.Buckets.Insert(self.ProjectId, &storage.Bucket{Location: self.Location, Name: self.BucketName}).Do(); err == nil {
			fmt.Printf("Created bucket %v at location %v\n\n", res.Name, res.SelfLink)
		} else {
			self.fatalf(service, "Failed creating bucket %s: %v\n\n", self.BucketName, err)
		}
	}
}

func (self *GoogleCloudStorage) copy(objectName string, destinationBucket string) string {
	client, err := google.DefaultClient(context.Background(), scope, compute.ComputeScope)
	if err != nil {
		log.Printf("Unable to get default client: %v", err)
	}
	service, err := storage.New(client)
	if err != nil {
		log.Printf("Unable to create storage service: %v", err)
	}

	object := &storage.Object{}
	if res, err := service.Objects.Copy(self.BucketName, objectName, destinationBucket, objectName, object).DestinationPredefinedAcl("publicRead").Do(); err == nil {
		fmt.Printf("Created object %v at location %v\n\n", res.Name, res.SelfLink)
		mediaLink := url.QueryEscape(res.MediaLink)
		return mediaLink
	} else {
		self.fatalf(service, "Objects.Copy failed: %v", err)
	}
	return ""
}

func (self *GoogleCloudStorage) move(objectName string, newName string, newBucket string) (string, string, error) {
	client, err := google.DefaultClient(context.Background(), scope, compute.ComputeScope)
	if err != nil {
		log.Printf("Unable to get default client: %v", err)
		return "", "", err
	}
	service, err := storage.New(client)
	if err != nil {
		log.Printf("Unable to create storage service: %v", err)
		return "", "", err
	}

	object := &storage.Object{}
	if res, err := service.Objects.Rewrite(self.BucketName, objectName, newBucket, newName, object).DestinationPredefinedAcl("publicRead").Do(); err == nil {
		fmt.Printf("Created object %v at location %v\n\n", res.Resource.Name, res.Resource.SelfLink)
		mediaLink := url.QueryEscape(res.Resource.MediaLink)
		data, _ := base64.StdEncoding.DecodeString(res.Resource.Md5Hash)
		return mediaLink, hex.EncodeToString(data), nil
	} else {
		log.Printf("Objects.Copy failed: %v", err)
	}
	return "", "", err
}

func (self *GoogleCloudStorage) GetMd5(objectName string) (string, string, error) {
	client, err := google.DefaultClient(context.Background(), scope, compute.ComputeScope)
	if err != nil {
		log.Printf("Unable to get default client: %v", err)
	}
	service, err := storage.New(client)
	if err != nil {
		log.Printf("Unable to create storage service: %v", err)
	}

	if res, err := service.Objects.Get(self.BucketName, objectName).Do(); err == nil {
		decode, _ := base64.StdEncoding.DecodeString(res.Md5Hash)
		return hex.EncodeToString(decode), res.MediaLink, err
	} else {
		log.Printf("Objects.Get failed: %v", err)
		return "", "", err
	}
}
