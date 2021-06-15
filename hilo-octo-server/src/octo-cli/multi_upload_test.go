package main

import (
	"octo-cli/utils"
	"github.com/codegangsta/cli"
	"testing"
)

// TODO API有必要启动。
func Setup() error {
	decodeTomle("../../config.tml")
	utils.AppSecret = "octo-app-secret"
	return nil
}

//
func TestMultiUploadAssetBundle(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}
	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadAssetBundle(9410,
		"/Users/a12889/Dropbox/gopath/src/hilo-octo-unity-test/Assets/StreamingAssets/asset_bundle/a/v1/v1.manifest",
		[]string{"init", app.Version}, 0, false, app.Version, false, "", "")

}

func TestMultiUploadAssetBundleWithCustomCors(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}
	Setup()

	app := cli.NewApp()
	app.Version = "v2.4"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadAssetBundle(100010,
		"/Users/a12889/Dropbox/gopath/src/hilo-octo-unity-test/Assets/StreamingAssets/asset_bundle/a/v1/v1.manifest",
		[]string{"init", app.Version}, 0, false, app.Version, true, `{"maxAge":60, "methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"], "origins": ["*"], "responseHeaders":["X-Octo-Key"]}`, "")

}

func TestMultiUploadAssetBundleUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}
	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadAssetBundle(9408,
		"",
		[]string{"update", app.Version}, 0, false, app.Version, false, "", "")
}

func TestMultiUploadOneAssetBundle(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}

	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadAssetBundle(9410,
		"/Users/a12889/Dropbox/gopath/src/hilo-octo-unity-test/Assets/StreamingAssets/asset_bundle/a/v1/v1.manifest",
		[]string{"one", app.Version}, 0, false, app.Version, false, "", "qe21507132")
}

func TestMultiUploadAllResources(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}

	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadResources(9403,
		"",
		[]string{"init", app.Version}, 0, false, app.Version, "", false, true, "")
}

func TestMultiUploadAllResourcesUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}

	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadResources(9332,
		"",
		[]string{"update", app.Version}, 0, false, app.Version, "", false, true, "")
}

func TestMultiUploadOneResources(t *testing.T) {
	if testing.Short() {
		t.Skip("skip testing in short mode ")
	}

	Setup()

	app := cli.NewApp()
	app.Version = "v2.1"
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	utils.AppSecret = "octo-app-secret"
	utils.App = app
	// TODO Set path AND VersionID
	MultiUploadResources(9405,
		"",
		[]string{"one", app.Version}, 0, false, app.Version, "", false, false, "")
}
