package main

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/utils"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/codegangsta/cli"
	"gopkg.in/matryer/try.v1"

	"time"
)

func uploadAllAssetBundle(versionId int, manifestPath string, tags cli.StringSlice, priority int, useOldTagFlg bool, buildNumber, corsStr string, cors bool) {

	if useOldTagFlg && len(tags) > 0 {
		log.Fatal("useOldTag and tags can not specified at the same time")
	}

	//Decode Manifest
	singleManifest := DecodeSingleManifest(manifestPath)
	basePath := filepath.Dir(manifestPath)

	log.Println("basePath:", basePath)

	const listUrlString = "%s/v1/upload/list/%d"
	listUrl := fmt.Sprintf(listUrlString, Conf.Api.BaseUrl, versionId)

	var m map[string]interface{}
	err := utils.HttpGet(listUrl, &m)
	if err != nil {
		utils.Fatal(err)
	}

	gcs := NewGoogleCloudStorage(m["ProjectId"].(string), m["Backet"].(string)+"-"+strconv.Itoa(versionId)+"-assetbundle", m["Location"].(string))
	gcs.createBucket()
	if cors {
		if corsStr == "" {
			corsV := GCPCORSValues{
				MaxAge:          60,
				Methods:         []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				Origins:         []string{"*"},
				ResponseHeaders: []string{"X-Octo-Key"},
			}
			gcs.setCORS(corsV)
		} else {
			gcs.setCORSWithJSON(corsStr)
		}
	}

	fileMap := m["Files"].(map[string]interface{})

	newFileList := []NewFile{}
	visitedMap := map[string]bool{}
	for name := range singleManifest.AssetBundleManifest {
		makeAssetBundleMap(&newFileList, visitedMap, gcs, fileMap, name, singleManifest, versionId, basePath, strings.Join(tags, ","), priority, buildNumber)
	}

	if len(newFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	jsonBytes, err := json.Marshal(newFileList)
	if err != nil {
		panic(err)
	}

	uploadUrlString := "%s/v1/upload/all/%d"
	if useOldTagFlg {
		uploadUrlString += "/notag"
	}
	uploadUrl := fmt.Sprintf(uploadUrlString, Conf.Api.BaseUrl, versionId)
	var res map[string]interface{}
	err = utils.HttpPost(uploadUrl, jsonBytes, &res)
	if err != nil {
		utils.Fatal(err)
	}
	log.Println("Upload Complete. RevisionId:", res["RevisionId"])
}

func makeAssetBundleMap(abList *[]NewFile, visitedMap map[string]bool, gcs *GoogleCloudStorage, fileMap map[string]interface{}, name string, singleManifest SingleManifest, versionId int, basePath string, tags string, priority int, buildNumber string) {
	dependencies := singleManifest.AssetBundleManifest[name].Dependencies

	if visitedMap[name] {
		if false {
			log.Println("[DEBUG] already visited:", name)
		}
		return
	}
	visitedMap[name] = true

	for _, dependencie := range dependencies {
		if !visitedMap[dependencie] {
			log.Println("Check dependencie:", dependencie)
			makeAssetBundleMap(abList, visitedMap, gcs, fileMap, dependencie, singleManifest, versionId, basePath, tags, priority, buildNumber)
		}
	}

	assetPath := filepath.FromSlash(basePath + "/" + name)
	manifest := DecodeManifest(assetPath + ".manifest")
	crc := manifest.CRC
	newFile := NewFile{
		Filename:    name,
		Dependency:  strings.Join(dependencies, ","),
		Priority:    priority,
		Tag:         tags,
		Crc:         crc,
		BuildNumber: buildNumber,
	}

	serverFileData := fileMap[name]
	if serverFileData == nil {
		log.Println(name, "is added.")

		const startUrlString = "%s/v1/upload/start?version=%d&filename=%s"
		startUrl := fmt.Sprintf(startUrlString, Conf.Api.BaseUrl, versionId, name)

		var m map[string]interface{}
		err := startUploadMetaDataWithRetry(startUrl, &m)

		if err != nil {
			utils.Fatal(err)
		}

		err = uploadFile(gcs, assetPath, m["FileName"].(string), newFile, abList)

		if err != nil {
			utils.Fatal(err)
		}
		return
	}

	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileCrc := uint32(serverFileDataM["CRC"].(float64))
	if serverFileCrc != crc || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		log.Println(name, "is changed.")
		log.Println("old CRC is", serverFileCrc)
		log.Println("new CRC is", crc)

		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("encrypted name is", encryptedName)
		err := uploadFile(gcs, assetPath, encryptedName, newFile, abList)
		if err != nil {
			utils.Fatal(err)
		}
		return
	}

	log.Println(name, "is nochanged.")
}

func uploadAllResources(versionId int, basePath string, tags cli.StringSlice, priority int, useOldTagFlg bool, buildNumber, corsStr string, cors bool, recursion bool) {

	if useOldTagFlg && len(tags) > 0 {
		log.Fatal("useOldTag and tags can not specified at the same time")
	}

	log.Println("basePath:", basePath)

	const listUrlString = "%s/v1/resource/upload/list/%d"
	listUrl := fmt.Sprintf(listUrlString, Conf.Api.BaseUrl, versionId)

	var m map[string]interface{}
	err := utils.HttpGet(listUrl, &m)
	if err != nil {
		utils.Fatal(err)
	}

	gcs := NewGoogleCloudStorage(m["ProjectId"].(string), m["Backet"].(string)+"-"+strconv.Itoa(versionId)+"-resources", m["Location"].(string))
	gcs.createBucket()
	if cors {
		if corsStr == "" {
			corsV := GCPCORSValues{
				MaxAge:          60,
				Methods:         []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				Origins:         []string{"*"},
				ResponseHeaders: []string{"X-Octo-Key"},
			}
			gcs.setCORS(corsV)
		} else {
			gcs.setCORSWithJSON(corsStr)
		}
	}

	fileMap := m["Files"].(map[string]interface{})

	//check directory
	fInfo, err := os.Stat(basePath)
	if err != nil {
		utils.Fatal(err)
	}
	if !fInfo.IsDir() {
		log.Fatal(basePath, "is not Directory.")
	}

	var pathMap = make(map[string]string)
	//get and make filepath list
	makeFilePaths(basePath, recursion, pathMap)

	newFileList := []NewFile{}

	for _, path := range pathMap {
		resourceMap(&newFileList, gcs, fileMap, versionId, path, strings.Join(tags, ","), priority, buildNumber)
	}

	if len(newFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	jsonBytes, err := json.Marshal(newFileList)
	if err != nil {
		panic(err)
	}

	uploadUrlString := "%s/v1/resource/upload/all/%d"
	if useOldTagFlg {
		uploadUrlString += "/notag"
	}
	uploadUrl := fmt.Sprintf(uploadUrlString, Conf.Api.BaseUrl, versionId)
	var res map[string]interface{}
	err = utils.HttpPost(uploadUrl, jsonBytes, &res)
	if err != nil {
		utils.Fatal(err)
	}
	log.Println("Upload Complete. RevisionId:", res["RevisionId"])
}

func makeFilePaths(fpath string, isRecursion bool, pathMap map[string]string) {
	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		log.Println("makeFileInfos read dir error")
		utils.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() && isRecursion {
			makeFilePaths(filepath.Join(fpath, file.Name()), isRecursion, pathMap)
			continue
		}
		pathMap[file.Name()] = filepath.Join(fpath, file.Name())
	}
}

func resourceMap(abList *[]NewFile, gcs *GoogleCloudStorage, fileMap map[string]interface{}, versionId int, path string, tags string, priority int, buildNumber string) {

	fileInfo, err := getFileInfo(path)
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir {
		return
	}

	newFile := NewFile{
		Filename:    fileInfo.Name,
		Priority:    priority,
		Tag:         tags,
		Size:        fileInfo.Size,
		MD5:         fileInfo.MD5,
		BuildNumber: buildNumber,
	}

	serverFileData := fileMap[fileInfo.Name]
	if serverFileData == nil {
		log.Println(fileInfo.Name, "is added.")
		//get new name
		const startUrlString = "%s/v1/resource/upload/start?version=%d&filename=%s"
		startUrl := fmt.Sprintf(startUrlString, Conf.Api.BaseUrl, versionId, fileInfo.Name)

		var m map[string]interface{}
		err := startUploadMetaDataWithRetry(startUrl, &m)
		if err != nil {
			utils.Fatal(err)
		}

		err = uploadFile(gcs, path, m["FileName"].(string), newFile, abList)

		if err != nil {
			utils.Fatal(err)
		}
		return
	}

	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileMD5 := serverFileDataM["MD5"].(string)
	if serverFileMD5 != newFile.MD5 || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		log.Println(fileInfo.Name, "is changed.")
		log.Println("old MD5 is", serverFileMD5)
		log.Println("new MD5 is", newFile.MD5)

		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("encrypted name is", encryptedName)
		err := uploadFile(gcs, path, encryptedName, newFile, abList)
		if err != nil {
			utils.Fatal(err)
		}
		return
	}

	log.Println(fileInfo.Name, "is nochanged.")

}

func uploadFile(gcs *GoogleCloudStorage, path string, fileName string, newFile NewFile, abList *[]NewFile) error {

	err := try.Do(func(attempt int) (bool, error) {
		mediaLink, err := gcs.upload(path, fileName)
		if err != nil {
			time.Sleep(5 * time.Second)
			return attempt < 3, err
		}
		newFile.Url = mediaLink
		newFile.EnciptName = fileName
		fileInfo, err := getFileInfo(path)
		if err != nil {
			log.Fatal(err)
		}
		newFile.Size = fileInfo.Size
		newFile.MD5 = fileInfo.MD5
		*abList = append(*abList, newFile)
		return false, nil
	})
	return err
}
