package main

import (
	"encoding/json"
	"fmt"
	"log"

	"hilo-octo-cli/src/octo-cli/utils"

	"github.com/codegangsta/cli"
)

func addTagToAssetBundle(versionId int, files cli.StringSlice, tags cli.StringSlice) {
	const urlString = "%s/v1/tag/ab/%d"
	addTag(versionId, files, tags, urlString)
}

func addTagToResource(versionId int, files cli.StringSlice, tags cli.StringSlice) {
	const urlString = "%s/v1/tag/r/%d"
	addTag(versionId, files, tags, urlString)
}

func addTag(versionId int, files cli.StringSlice, tags cli.StringSlice, urlString string) {

	type Rec struct {
		Files cli.StringSlice
		Tags  cli.StringSlice
	}

	rec := Rec{
		Files: files,
		Tags:  tags,
	}

	jsonBytes, err := json.Marshal(rec)
	if err != nil {
		panic(err)
	}
	log.Println(string(jsonBytes))

	url := fmt.Sprintf(urlString, Conf.Api.BaseUrl, versionId)
	if err := utils.HttpPost(url, jsonBytes, nil); err != nil {
		utils.Fatal(err)
	}
}

func removeTagToAssetBundle(versionId int, files cli.StringSlice, tags cli.StringSlice) {
	const urlString = "%s/v1/remove/tag/ab/%d"
	removeTag(versionId, files, tags, urlString)
}

func removeTagToResource(versionId int, files cli.StringSlice, tags cli.StringSlice) {
	const urlString = "%s/v1/remove/tag/r/%d"
	removeTag(versionId, files, tags, urlString)
}

func removeTag(versionId int, files cli.StringSlice, tags cli.StringSlice, urlString string) {

	type Rec struct {
		Files cli.StringSlice
		Tags  cli.StringSlice
	}

	rec := Rec{
		Files: files,
		Tags:  tags,
	}

	jsonBytes, err := json.Marshal(rec)
	if err != nil {
		panic(err)
	}
	log.Println(string(jsonBytes))

	url := fmt.Sprintf(urlString, Conf.Api.BaseUrl, versionId)
	if err := utils.HttpPost(url, jsonBytes, nil); err != nil {
		utils.Fatal(err)
	}
}
