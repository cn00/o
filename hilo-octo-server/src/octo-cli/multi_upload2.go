package main

import (
	"github.com/codegangsta/cli"
	"io/ioutil"
	"log"
	"octo-cli/utils"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type IUploadSupplier interface {
	startUploadToOSS(priority int, tags cli.StringSlice, buildNumber, uploadType string, fileMap *FileMap) int
}

var ossRoCos = "cos"

func UploadAssetBundle(versionId int, manifestPath string, tags cli.StringSlice, priority int, useOldTagFlg bool,
	buildNumber string, cors bool, corsStr, specificManifest, filter string, c *cli.Context) {

	listMode := c.Bool("list")
	if listMode {

	}
	ossRoCos = c.String("supplier")

	log.Println("UploadAssetBundle：", manifestPath, ossRoCos)
	mf, _ := os.Stat(manifestPath)
	if mf.IsDir() {
		ncpu := runtime.NumCPU()
		log.Println("遍历文件夹：", manifestPath, "ncpu:", ncpu)
		manifests, _ := ioutil.ReadDir(manifestPath)
		manifestList := []string{}
		for _, mi := range manifests {
			if !mi.IsDir() && strings.HasSuffix(mi.Name(), ".manifest") {

				if len(filter) > 0 {
					match, _ := regexp.MatchString(filter, mi.Name())
					if !match {
						continue
					}
				}
				//manifestList = append(manifestList, mi.Name())
				manifestList = append(manifestList, manifestPath+"/"+mi.Name())
				//doOneManifestfile(versionId , manifestPath+"/"+mi.Name() , tags , priority , useOldTagFlg , buildNumber , cors , corsStr, specificManifest )
			}
		}
		count := len(manifestList)
		log.Println("manifestList", count)

		doManyManifestfile(versionId, manifestList, tags, priority, useOldTagFlg, buildNumber, cors, corsStr, specificManifest)

	} else {
		log.Println("单文件：", manifestPath)
		doManyManifestfile(versionId, []string{manifestPath}, tags, priority, useOldTagFlg, buildNumber, cors, corsStr, specificManifest)
	}

}

func doManyManifestfile(versionId int, manifestPath []string, tags cli.StringSlice, priority int, useOldTagFlg bool,
	buildNumber string, cors bool, corsStr, specificManifest string) {

	start := time.Now()

	//singleManifest := DecodeManifest(manifestPath)
	singleManifest := DecodeBundleManifests(manifestPath)
	if singleManifest != nil {
		//err := try.Do(func(attempt int) (retry bool, err error) {
		var fileMap = FileMap{}
		fileMap.option.tag = strings.Join(tags, ",")
		fileMap.option.priority = priority
		fileMap.option.buildNumber = buildNumber
		fileMap.uploadedFileList = []NewFile{}
		fileMap.objectNameMap = map[string]string{}
		fileMap.planUploadFileMap = map[string]GCSFile{}
		fileMap.serverFileMap = map[string]interface{}{}
		fileMap.fileMap = map[string]interface{}{}
		fileMap.pathMap = map[string]string{}
		fileMap.ErrorFileMap = map[string]string{}
		fileMap.option = UploadOption{}

		err := getListOnServer(versionId, AssetBundleListURLPath, &fileMap)
		if err != nil {
			utils.Fatal(err)
		}

		basePath := filepath.Dir(manifestPath[0])

		// 创建上传信息
		createAssetBundleUploadInfo(*singleManifest, basePath, specificManifest, versionId, &fileMap)

		depsChangedCount := len(fileMap.uploadedFileList)

		// OSS 上传
		var oss IUploadSupplier
		if ossRoCos == "cos" {
			oss = NewTencentCOS()
		} else if ossRoCos == "oss" {
			oss = NewAliyunOSS()
		} else if ossRoCos == "gcs" {
			oss = NewGoogleCloud()
		}
		assetCount := oss.startUploadToOSS(priority, tags, buildNumber, UploadTypeAssetBundle, &fileMap)

		// 上传结果OCTO API发送到
		endUpload(useOldTagFlg, versionId, AssetBundleUploadAllURLPath, &fileMap)

		elapsed := time.Since(start)
		log.Printf("Uploaded AssetBundle count: %d, Deps changed count: %d, Elapsed Time: %f Seconds\n", assetCount, depsChangedCount, elapsed.Seconds())

		printErrorFile()
		//return false, nil
		//})
		//log.Println("doOneManifestfile", manifestPath, err)
	}
}

func DecodeBundleManifests(manifestFiles []string) *SingleManifest {
	manifest := new(SingleManifest)
	var assetBundleMap = map[string]AssetBundleInfo{}
	for _, manifestFile := range manifestFiles {
		manifest1 := DecodeManifest(manifestFile)

		manifest.ManifestFileVersion = manifest1.ManifestFileVersion

		assetName := manifestFile[strings.LastIndex(manifestFile, "/")+1 : strings.LastIndex(manifestFile, ".")]
		assetBundleMap[assetName] = AssetBundleInfo{Dependencies: manifest1.Dependencies, CRC: manifest1.CRC}
	}
	manifest.AssetBundleManifest = assetBundleMap
	return manifest
}
