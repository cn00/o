package list

import (
	"fmt"
	"log"
	"strings"

	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/config"
	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/utils"
)

type ListOptions struct {
	Config    config.Config
	VersionId int
}

func ListAssetbundles(o ListOptions) {
	list(o, "ab")
}

func ListResources(o ListOptions) {
	list(o, "r")
}

func list(o ListOptions, t string) {
	url := fmt.Sprint(o.Config.Api.BaseUrl, "/v2/admin/a/", o.Config.App.Id, "/list/", t, "/", o.VersionId)
	var fileList []struct {
		Filename string   `json:"filename"`
		Tag      []string `json:"tag"`
	}
	err := utils.HttpGet(url, &fileList)
	if err != nil {
		utils.Fatal(err)
	}

	w := utils.NewTabwriter()
	fmt.Fprintln(w, "name\ttags")
	for _, file := range fileList {
		fmt.Fprint(w, file.Filename, "\t", strings.Join(file.Tag, ","), "\n")
	}
	w.Flush()

	log.Printf("Total: %d", len(fileList))
}
