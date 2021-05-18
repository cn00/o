package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/utils"
)

func checkAssetBundleExistence(versionId int, files string) {
	checkExistence(versionId, files, "ab")
}

func checkResourceExistence(versionId int, files string) {
	checkExistence(versionId, files, "r")
}

func checkExistence(versionId int, files string, t string) {
	url := fmt.Sprint(Conf.Api.BaseUrl, "/v2/admin/a/", Conf.App.Id, "/list/", t, "/", versionId)
	var fileList []struct {
		Filename string   `json:"filename"`
		Tag      []string `json:"tag"`
	}
	err := utils.HttpGet(url, &fileList)
	if err != nil {
		utils.Fatal(err)
	}

	var filenames []string
	if files != "" {
		filenames = strings.Split(files, ",")
	}

	fileMap := make(map[string]bool, len(fileList))
	for _, file := range fileList {
		fileMap[file.Filename] = true
	}

	var yesCount, noCount int
	w := utils.NewTabwriter()
	fmt.Fprintln(w, "name\texist")
	for _, filename := range filenames {
		if fileMap[filename] {
			yesCount++
			fmt.Fprint(w, filename)
			fmt.Fprintln(w, "\tyes\t")
			continue
		}
		noCount++
		fmt.Fprint(w, filename)
		fmt.Fprintln(w, "\tno\t")
	}
	w.Flush()

	log.Printf("Exist:%d, Not exist:%d", yesCount, noCount)

	if noCount > 0 {
		log.Fatal("Nonexistent files found")
	}
}
