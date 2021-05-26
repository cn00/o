package main

import (
	"encoding/json"
	"fmt"
	"log"

	"hilo-octo-cli/src/octo-cli/utils"

	"github.com/codegangsta/cli"
)

func deleteAssetBundle(versionId int, files cli.StringSlice) {
	const urlString = "%s/v1/delete/ab/%d"
	delete(versionId, files, urlString)
}

func deleteResource(versionId int, files cli.StringSlice) {
	const urlString = "%s/v1/delete/r/%d"
	delete(versionId, files, urlString)
}

func delete(versionId int, files cli.StringSlice, urlString string) {

	type Rec struct {
		Files cli.StringSlice
	}

	rec := Rec{
		Files: files,
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
