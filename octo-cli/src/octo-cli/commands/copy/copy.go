package copy

import (
	"encoding/json"
	"fmt"

	"log"

	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/config"
	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/utils"
)

type CopyOptions struct {
	Config               config.Config
	SourceVersionId      int
	DestinationVersionId int
	Filenames            []string
	Debug                bool
}

func CopyAssetBundle(o CopyOptions) {
	copy(o, "ab")
}

func CopyResource(o CopyOptions) {
	copy(o, "r")
}

func copy(o CopyOptions, t string) {
	req := struct {
		SourceVersionId      int      `json:"source_version_id"`
		DestinationVersionId int      `json:"destination_version_id"`
		Filenames            []string `json:"filenames"`
	}{
		SourceVersionId:      o.SourceVersionId,
		DestinationVersionId: o.DestinationVersionId,
		Filenames:            o.Filenames,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	if o.Debug {
		log.Println("[DEBUG] reqBytes:", string(reqBytes))
	}
	url := fmt.Sprint(o.Config.Api.BaseUrl, "/v2/admin/a/", o.Config.App.Id, "/copy/", t)
	var res interface{}
	err = utils.HttpPost(url, reqBytes, &res)
	if err != nil {
		utils.Fatal(err)
	}

	resJSONBytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	log.Println("[INFO] Copy succeeded.")
	log.Printf("[INFO] %s\n", resJSONBytes)
}
