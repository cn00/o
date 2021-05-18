package main

import (
	"cloud.google.com/go/storage"
	"context"
	"log"
	"testing"
	//"src/github.com/stretchr/testify/assert"
)

func TestCreateBucket(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}

	gcs := NewGoogleCloudStorage("hilo-1047", "golangtest-9680-assetbundle", "asia")
	gcs.createBucket()
	ctx := context.Background()

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Get a Bucket instance.
	bucket := client.Bucket(gcs.BucketName)

	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !attrs.VersioningEnabled {
		log.Fatal("Versioning is disabled!")
	}
	expectedRole := []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if attrs.DefaultObjectACL[0] != expectedRole[0] {
		log.Fatal("DefaultObjectACL is not equal!")
	}
}

func TestUpdateCORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode")
	}
	gcs := NewGoogleCloudStorage("hilo-1047", "golangtest-9680-assetbundle", "asia")
	corsValue := GCPCORSValues{
		MaxAge:          60,
		Methods:         []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		Origins:         []string{"*"},
		ResponseHeaders: []string{"X-Octo-Key"},
	}
	gcs.setCORS(corsValue)

	ctx := context.Background()

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Get a Bucket instance.
	bucket := client.Bucket(gcs.BucketName)

	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if attrs.CORS == nil {
		log.Fatal("Fail update CORS!")
	}

	expectedCORS := []storage.CORS{{MaxAge: 60, Methods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, Origins: []string{"*"}, ResponseHeaders: []string{"X-Octo-Key"}}}

	if attrs.CORS[0].ResponseHeaders[0] != expectedCORS[0].ResponseHeaders[0] {
		log.Fatal("Not Equal between actual CORS ResponseHeaders and expected CORS ResponseHeaders")
	}

	if len(attrs.CORS[0].Methods) != len(expectedCORS[0].Methods) {
		log.Fatal("Not Equal between actual CORS Methods and expected CORS Methods")
	}
}

func TestUpdateCustomCORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode")
	}
	gcs := NewGoogleCloudStorage("hilo-1047", "golangtest-9680-assetbundle", "asia")

	jsonCors := `{"maxAge":60, "methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"], "origins": ["*"], "responseHeaders":["X-Octo-Key"]}`
	gcs.setCORSWithJSON(jsonCors)

	ctx := context.Background()

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Get a Bucket instance.
	bucket := client.Bucket(gcs.BucketName)

	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if attrs.CORS == nil {
		log.Fatal("Fail update CORS!")
	}

	expectedCORS := []storage.CORS{{MaxAge: 60, Methods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, Origins: []string{"*"}, ResponseHeaders: []string{"X-Octo-Key"}}}

	if attrs.CORS[0].ResponseHeaders[0] != expectedCORS[0].ResponseHeaders[0] {
		log.Fatal("Not Equal between actual CORS ResponseHeaders and expected CORS ResponseHeaders")
	}

	if len(attrs.CORS[0].Methods) != len(expectedCORS[0].Methods) {
		log.Fatal("Not Equal between actual CORS Methods and expected CORS Methods")
	}
}
