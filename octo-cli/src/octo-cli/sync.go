package main

import (
	"fmt"
	"log"

	"hilo-octo-cli/src/octo-cli/utils"
)

func diffSync(versionId int, sourceAppId int, sourceVersionId int, revisionId int) {
	const urlString = "%s/v1/sync/%d/%d/diff/%d/%d"
	url := fmt.Sprintf(urlString, Conf.Api.BaseUrl, versionId, revisionId, sourceAppId, sourceVersionId)
	sync(url)
}

func diffSyncLatest(versionId int, sourceAppId int, sourceVersionId int) {
	const urlString = "%s/v1/sync/%d/latest/diff/%d/%d"
	url := fmt.Sprintf(urlString, Conf.Api.BaseUrl, versionId, sourceAppId, sourceVersionId)
	sync(url)
}

func sync(url string) {
	err := utils.HttpPost(url, nil, nil)
	if err != nil {
		utils.Fatal(err)
	}
	log.Println("finish sync.")
}
