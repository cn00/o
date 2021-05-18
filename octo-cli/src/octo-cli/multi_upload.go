package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/QualiArts/hilo-octo-cli/src/octo-cli/utils"
	"github.com/QualiArts/hilo-octo-proto/go/octo"
	"github.com/codegangsta/cli"
	"github.com/pkg/errors"
	"gopkg.in/matryer/try.v1"
	"io"
	"io/ioutil"
	"log"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	sync2 "sync"
	"time"
)

type UploadOption struct {
	tag string
	priority int
	buildNumber string
}

type File struct {
	Id  int
	CRC uint32
}

// NewFile 新しいファイル
type NewFile struct {
	Filename    string
	EnciptName  string
	Size        int64
	Crc         uint32
	MD5         string
	Priority    int
	Tag         string
	Assets      []string
	Dependency  string
	Url         string
	BuildNumber string
}

// FileInfo ファイル情報
type FileInfo struct {
	Size  int64
	MD5   string
	Name  string
	IsDir bool
}

// GCSFile GCSにファイルアップロード後、取得情報
type GCSFile struct {
	FilePath        string
	FileName        string
	AssetBundleName string
	ObjectName      string
	MediaLink       string
	Dependencies    []string
	Err             error
}

const (
	retryInterval = 15 * time.Second

	// UploadTypeResources Resourcesのアップロードタイプ
	UploadTypeResources = "resources"
	// UploadTypeResources AssetBundleのアップロードタイプ
	UploadTypeAssetBundle = "assetbundle"

	// UpdateJudgeStrCRC CRCはAssetBundleの更新確認で利用
	UpdateJudgeStrCRC = "CRC"
	// UpdateJudgeStrMD5 MD5はResourcesの更新確認で利用
	UpdateJudgeStrMD5 = "MD5"

	// AssetBundleUploadAllURLPath AssetBundleアップロード最後の処理で利用するパス
	AssetBundleUploadAllURLPath = "%s/v1/upload/all/%d"
	// ResourcesUploadAllURLPath Resourcesアップロード最後の処理で利用するパス
	ResourcesUploadAllURLPath = "%s/v1/resource/upload/all/%d"

	// AssetBundleListURLPath サーバー上にアップロードされているAssetBundleの情報を取得するパス
	AssetBundleListURLPath = "%s/v1/upload/list/%d"
	// ResourcesListURLPath サーバー上にアップロードされているResourcesの情報を取得するパス
	ResourcesListURLPath = "%s/v1/resource/upload/list/%d"

	// AssetBundleUploadStartURLPath AssetBundleアップロード開始のパス
	AssetBundleUploadStartURLPath = "%s/v1/upload/start?version=%d&filename=%s"
	// ResourcesUploadStartURLPath Resourcesアップロード開始のパス
	ResourcesUploadStartURLPath = "%s/v1/resource/upload/start?version=%d&filename=%s"

	// MaxOutStanding fileUploadで利用するchan bufferの数
	MaxOutStanding = 100
)

var (
	uploadedFileList    []NewFile
	uploadUrlVisitedMap = map[string]bool{}
	objectNameMap       = map[string]string{}
	planUploadFileMap   = map[string]GCSFile{}
	serverFileMap       = map[string]interface{}{}
	fileMap             = map[string]interface{}{}
	pathMap             = map[string]string{}
	ErrorFileMap        = map[string]string{}

	option UploadOption
)

// AssetBundleアップロード処理順序
// 1. prepareUploadでGCSの情報を取得、bucketも作成
// 2. AssetBundleの確認
//    特定AssetBundleのみ登録の確認
//    アップロードで必要な情報を作成（新規かアップデートか）
//    OCTOサーバー上のファイル確認APIとStart APIを実行
// 3. 2の情報元にGCSにアップロード
// 4. アップロード終了しOCTOのALL APIを実行
// MultiUploadAssetBundle AssetBundleアップロード（並列処理）
func MultiUploadAssetBundle(versionId int, manifestPath string, tags cli.StringSlice, priority int, useOldTagFlg bool,
	buildNumber string, cors bool, corsStr, specificManifest string) {

	option.tag = strings.Join(tags, ",")
	option.priority = priority
	option.buildNumber = buildNumber

	start := time.Now()

	gcs := prepareUpload(versionId, AssetBundleListURLPath, cors, corsStr, UploadTypeAssetBundle)

	singleManifest := DecodeSingleManifest(manifestPath)
	basePath := filepath.Dir(manifestPath)

	// アップロード情報作成
	createAssetBundleUploadInfo(singleManifest, basePath, specificManifest, versionId)

	depsChangedCount := len(uploadedFileList)

	// GCPのGCSにアップロード
	assetCount := startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeAssetBundle)

	// アップロード結果をOCTO APIに送信
	endUpload(useOldTagFlg, versionId, AssetBundleUploadAllURLPath)

	elapsed := time.Since(start)
	log.Printf("Uploaded AssetBundle count: %d, Deps changed count: %d, Elapsed Time: %f Seconds\n", assetCount, depsChangedCount, elapsed.Seconds())

	printErrorFile()
}

// Resourceアップロード処理順序
// 1. prepareUploadでGCSの情報を取得、bucketも作成
// 2. 特定ファイルのみアップロードの場合ファイル確認
//    多数のファイルアップロードの場合Directoryの確認
//    アップロードで必要な情報を作成（新規かアップデートか）
//    OCTOサーバー上のファイル確認APIとStart APIを実行
// 3. 2の情報元にGCSにアップロード
// 4. アップロード終了しOCTOのALL APIを実行
// MultiUploadResources Resourcesアップロード（並列処理）
func MultiUploadResources(versionId int, basePath string, tags cli.StringSlice, priority int, useOldTagFlg bool, buildNumber, corsStr string, cors, recursion bool, specificFilePath string) {

	if useOldTagFlg && len(tags) > 0 {
		log.Fatal("useOldTag and tags can not specified at the same time")
	}

	option.tag = strings.Join(tags, ",")
	option.priority = priority
	option.buildNumber = buildNumber

	start := time.Now()

	gcs := prepareUpload(versionId, ResourcesListURLPath, cors, corsStr, UploadTypeResources)

	// 単独であげたい場合
	if len(specificFilePath) > 0 {
		multiUploadOneResource(versionId, tags, priority, useOldTagFlg, buildNumber, gcs, start, specificFilePath)
		return
	}

	//check directory
	checkDir(basePath)

	createResourceUploadInfo(basePath, versionId, recursion)

	resourcesCount := startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeResources)

	if len(uploadedFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	endUpload(useOldTagFlg, versionId, ResourcesUploadAllURLPath)

	elapsed := time.Since(start)
	log.Printf("Upload Resources count: %d, Elapsed Time: %f Seconds\n", resourcesCount, elapsed.Seconds())

	printErrorFile()
}

func createAssetBundleUploadInfo(singleManifest SingleManifest, basePath string, specificManifest string, versionId int) {
	// アップロードする前にstart apiを叩き、object mapを作成する
	createAssetBundleUploadStartFileMapAndObjectMap(singleManifest, basePath, specificManifest, versionId)
	for manifest := range uploadUrlVisitedMap {
		createUploadAssetBundle(manifest, singleManifest, basePath)
	}
}

func createResourceUploadInfo(basePath string, versionId int, recursion bool) {

	createResourcesUploadStartFileMap(basePath, versionId, recursion)

	for _, path := range pathMap {
		fileInfo, err := getFileInfo(path)
		if err != nil {
			log.Fatal(err)
		}

		if fileInfo.IsDir {
			continue
		}
		if len(fileInfo.Name) == 0 {
			continue
		}
		serverFileData := fileMap[fileInfo.Name]
		if serverFileData == nil {
			objectName := objectNameMap[fileInfo.Name]
			setUploadNewFileMap(path, fileInfo.Name, objectName, nil)
		} else {
			checkUpdateResourceFile(path, serverFileData)
		}
	}
}

func prepareUpload(versionId int, listUrlString string, cors bool, corsStr, uploadType string) *GoogleCloudStorage {

	// サーバー上の既存のファイルリストを取得する
	err := getListOnServer(versionId, listUrlString)
	if err != nil {
		utils.Fatal(err)
	}

	// gcs上bucket作成とcors設定
	gcs := createBucket(versionId, cors, corsStr, uploadType)
	return gcs
}

func multiUploadOneResource(versionId int, tags cli.StringSlice, priority int, useOldTagFlg bool, buildNumber string, gcs *GoogleCloudStorage, start time.Time, specificFilePath string) {

	//check file
	fileInfo := checkFile(specificFilePath)

	createOneResourceUploadInfo(fileInfo, versionId, specificFilePath)

	resourcesCount := startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeResources)

	endUpload(useOldTagFlg, versionId, ResourcesUploadAllURLPath)

	elapsed := time.Since(start)
	log.Printf("Upload Resources count: %d, Elapsed Time: %f Seconds\n", resourcesCount, elapsed.Seconds())

	printErrorFile()
}

func createOneResourceUploadInfo(fileInfo FileInfo, versionId int, specificFilePath string) {
	serverFileData := fileMap[fileInfo.Name]
	judgeNewOrUpdate(serverFileData, versionId, fileInfo.Name, specificFilePath, ResourcesUploadStartURLPath, UpdateJudgeStrMD5, nil, fileInfo)
	if serverFileData == nil {
		on := objectNameMap[fileInfo.Name]
		setUploadNewFileMap(specificFilePath, fileInfo.Name, on, nil)
	} else {
		checkUpdateResourceFile(specificFilePath, serverFileData)
	}
}

func createUploadAssetBundle(manifest string, singleManifest SingleManifest, basePath string) {

	ds := singleManifest.AssetBundleManifest[manifest].Dependencies

	ap := filepath.FromSlash(basePath + "/" + manifest)
	data := fileMap[manifest]
	if data == nil {
		on := objectNameMap[manifest]
		setUploadNewFileMap(ap, manifest, on, ds)
	} else {
		// 既存にあるファイルの場合はそのファイルをチェックする
		checkUpdateAssetBundleFile(ap, data, manifest, ds)
	}
}

func checkUpdateAssetBundleFile(assetPath string, serverFileData interface{}, fileName string, dependency []string) {
	manifest := DecodeManifest(assetPath + ".manifest")
	crc := manifest.CRC
	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileCrc := uint32(serverFileDataM["CRC"].(float64))
	if serverFileCrc != crc || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		log.Println(fileName, "is changed.")
		log.Println("old CRC is", serverFileCrc)
		log.Println("new CRC is", crc)
		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("encrypted name is", encryptedName)
		setUploadNewFileMap(assetPath, fileName, encryptedName, dependency)
	}
}

func checkUpdateResourceFile(assetPath string, serverFileData interface{}) {
	fileInfo, err := getFileInfo(assetPath)
	if err != nil {
	}
	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileMD5 := serverFileDataM["MD5"].(string)
	if serverFileMD5 != fileInfo.MD5 || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		log.Println(fileInfo.Name, "is changed.")
		log.Println("old MD5 is", serverFileMD5)
		log.Println("new MD5 is", fileInfo.MD5)

		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("encrypted name is", encryptedName)
		log.Println("assetPath is", assetPath)
		setUploadNewFileMap(assetPath, fileInfo.Name, encryptedName, nil)
	}
}

func setUploadNewFileMap(filePath string, fileName string, objectName string, dependencies []string) {
	if len(objectName) > 0 {
		planUploadFileMap[fileName] = GCSFile{
			FilePath: filePath,
			FileName: fileName,
			// AssetBundleの場合こちらにもセット
			AssetBundleName: fileName,
			ObjectName:      objectName,
			Dependencies:    dependencies,
		}
	}
}

func startUploadToGCP(gcs *GoogleCloudStorage, priority int, tags cli.StringSlice, buildNumber, uploadType string) int {
	var wg sync2.WaitGroup
	fileChan := make(chan GCSFile, MaxOutStanding)
	sem := make(chan struct{}, MaxOutStanding)
	count := len(planUploadFileMap)

	log.Println(uploadType+"Count", count)

	wg.Add(count)
	for _, v := range planUploadFileMap {

		go func(f GCSFile) {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()
			fileChan <- gcs.uploadWithChan(f.FilePath, f.FileName, f.AssetBundleName, f.ObjectName, f.Dependencies)
		}(v)
	}

	errorCount := 0
	for range planUploadFileMap {
		gcsFile := <-fileChan
		if gcsFile.Err == nil {
			createUploadedNewFileList(gcsFile, priority, tags, buildNumber, uploadType)
		} else {
			log.Printf("Upload error of %s: %+v", gcsFile.AssetBundleName, gcsFile.Err)
			errorCount += 1
		}
	}
	wg.Wait()
	close(sem)
	if errorCount > 0 {
		// アップロードエラーが1つ以上あったので中止
		log.Fatalln("Had error while uploading and cancel commit")
	}
	return count

}

func createUploadedNewFileList(gcsFile GCSFile, priority int, tags cli.StringSlice, buildNumber, uploadType string) {
	fileInfo, err := getFileInfo(gcsFile.FilePath)
	if err != nil {
		utils.Fatal(err)
	}

	var nFile NewFile
	if uploadType == UploadTypeResources {
		nFile = newFile(gcsFile, priority, tags, fileInfo, 0, buildNumber)
	} else if uploadType == UploadTypeAssetBundle {
		nFile = newFileAssetBundle(gcsFile, priority, tags, fileInfo, 0, buildNumber, gcsFile.AssetBundleName)
	}
	serverFileData := fileMap[gcsFile.FileName]
	if serverFileData == nil {
		if uploadType == UploadTypeAssetBundle {
			manifest := DecodeManifest(gcsFile.FilePath + ".manifest")
			crc := manifest.CRC
			nFile.Crc = crc
			nFile.Assets = manifest.Assets
		}
		uploadedFileList = append(uploadedFileList, nFile)
	} else {
		if uploadType == UploadTypeResources {
			serverFileDataM := serverFileData.(map[string]interface{})
			serverFileMD5 := serverFileDataM["MD5"].(string)
			if serverFileMD5 != nFile.MD5 || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
				nFile.EnciptName = serverFileDataM["EncriptedName"].(string)
				uploadedFileList = append(uploadedFileList, nFile)
			}
		} else if uploadType == UploadTypeAssetBundle {
			manifest := DecodeManifest(gcsFile.FilePath + ".manifest")
			crc := manifest.CRC
			nFile.Crc = crc
			nFile.Assets = manifest.Assets
			serverFileDataM := serverFileData.(map[string]interface{})
			serverFileCrc := uint32(serverFileDataM["CRC"].(float64))
			// TODO MD5比較も追加する
			if serverFileCrc != crc || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
				encryptedName := serverFileDataM["EncriptedName"].(string)
				nFile.EnciptName = encryptedName
				uploadedFileList = append(uploadedFileList, nFile)
			}
		}

	}
}

func newFile(v2 GCSFile, priority int, tags cli.StringSlice, fileInfo FileInfo, crc uint32, buildNumber string) NewFile {

	newFile := createNewFile(v2, priority, tags, fileInfo, crc, buildNumber, fileInfo.Name)

	return newFile
}

func newFileAssetBundle(v2 GCSFile, priority int, tags cli.StringSlice, fileInfo FileInfo, crc uint32, buildNumber string, assetBundleName string) NewFile {
	// AsseBundleNameはPathが含まれる可能性があるので、それをFileNameとして利用する
	newFile := createNewFile(v2, priority, tags, fileInfo, crc, buildNumber, assetBundleName)

	if len(v2.Dependencies) > 0 {
		newFile.Dependency = strings.Join(v2.Dependencies, ",")
	}

	return newFile
}

func createNewFile(v2 GCSFile, priority int, tags cli.StringSlice, fileInfo FileInfo, crc uint32, buildNumber, fileName string) NewFile {

	newFile := NewFile{
		Priority:    priority,
		Tag:         strings.Join(tags, ","),
		BuildNumber: buildNumber,
		EnciptName:  v2.ObjectName,
		Url:         v2.MediaLink,
	}
	if crc != 0 {
		newFile.Crc = crc
	}
	if fileInfo.MD5 != "" {
		newFile.MD5 = fileInfo.MD5
		newFile.Filename = fileName
		newFile.Size = fileInfo.Size
	}

	return newFile
}

func endUpload(useOldTagFlg bool, versionId int, uploadUrlString string) {
	if len(uploadedFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	jsonBytes, err := json.Marshal(uploadedFileList)
	if err != nil {
		panic(err)
	}

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

func createAssetBundleUploadStartFileMapAndObjectMap(singleManifest SingleManifest, basePath, specificManifest string, versionId int) (SingleManifest, string) {

	// SingleManifestの中のファイル名に半角スペースがあるかチェック
	for name := range singleManifest.AssetBundleManifest {
		checkFileNameWithSpace(name)
	}

	// ファイル名に半角がある場合、そのリストを出力して終了
	printErrorFile()

	// SingleManifestの中の一つのAssetのみ指定した場合
	if specificManifest != "" {
		if _, ok := singleManifest.AssetBundleManifest[specificManifest]; ok {
			uploadUrlVisitedMap[specificManifest] = true
			var dependencies = singleManifest.AssetBundleManifest[specificManifest].Dependencies
			assetPath := filepath.FromSlash(basePath + string(os.PathSeparator) + specificManifest)

			// 新しいファイルor更新ファイル判断
			judgeNewOrUpdate(fileMap[specificManifest], versionId, specificManifest, assetPath, AssetBundleUploadStartURLPath, UpdateJudgeStrCRC, dependencies, FileInfo{})
		}
	} else {
		// 特定ファイル指定ではない場合依存関係を再帰でチェックする
		for name := range singleManifest.AssetBundleManifest {
			checkDependencyAssetBundle(singleManifest, name, basePath, versionId)
		}
	}
	return singleManifest, basePath
}

func judgeNewOrUpdate(serverFileData interface{}, versionId int, name, path, startUrlString, checkStr string, dependencies []string, fileInfo FileInfo) {

	if serverFileData == nil {
		// new
		log.Println(name, "is added.")

		var fileStartMap map[string]interface{}

		startURL := fmt.Sprintf(startUrlString, Conf.Api.BaseUrl, versionId, name)
		err := startUploadMetaDataWithRetry(startURL, &fileStartMap)
		if err != nil {
			errorMsg := fmt.Sprintf("'%v' have a problem with err : %v", name, err)
			ErrorFileMap[name] = errorMsg
		}
		planUploadFileMap[name] = GCSFile{
			FilePath:     path,
			FileName:     name,
			ObjectName:   fileStartMap["FileName"].(string),
			Dependencies: dependencies,
		}
		objectNameMap[name] = fileStartMap["FileName"].(string)
		return
	}

	// update
	serverFileDataM := serverFileData.(map[string]interface{})
	var actual, expect interface{}
	if checkStr == UpdateJudgeStrCRC {
		manifest := DecodeManifest(path + ".manifest")
		actual = manifest.CRC
		expect = uint32(serverFileDataM[UpdateJudgeStrCRC].(float64))
	} else if checkStr == UpdateJudgeStrMD5 {
		actual = fileInfo.MD5
		expect = serverFileDataM[UpdateJudgeStrMD5].(string)
	}
	encryptedName := serverFileDataM["EncriptedName"].(string)
	state := octo.Data_State(serverFileDataM["State"].(float64))
	if expect != actual || state == octo.Data_DELETE {
		// had change in file
		planUploadFileMap[name] = GCSFile{
			FilePath:     path,
			FileName:     name,
			ObjectName:   encryptedName,
			Dependencies: nil,
		}
		return
	}

	// check deps, but old server does not support
	isDepsChanged := func() bool {
		if deps, ok := serverFileDataM["Dependencies"]; ok {
			if deps == nil {
				deps = []interface{}{}
			}
			serverDeps := deps.([]interface{})
			if len(serverDeps) != len(dependencies) {
				return true
			}
			// dont care calc cost
			for _, sd := range serverDeps {
				contains := false
				for _, d := range dependencies {
					if sd.(string) == d {
						contains = true
						break
					}
				}
				if !contains {
					return true
				}
			}
		}
		return false
	}()
	if isDepsChanged {
		log.Println(name, ": dependencies changed.")
		file := NewFile{
			Filename: name,
			Dependency: strings.Join(dependencies, ","),
			Tag: option.tag,
			Priority: option.priority,
			BuildNumber: option.buildNumber,
		}
		uploadedFileList = append(uploadedFileList, file)
		return
	}

	log.Println(name, "is no changed.")
}

func checkDependencyAssetBundle(singleManifest SingleManifest, name string, basePath string, versionId int) {
	var dependencies = singleManifest.AssetBundleManifest[name].Dependencies
	if uploadUrlVisitedMap[name] {
		if false {
			log.Println("[DEBUG] already visited:", name)
		}
	}
	uploadUrlVisitedMap[name] = true

	// 依存関係のチェック
	for _, dependency := range dependencies {
		if !uploadUrlVisitedMap[dependency] {
			log.Println(name+" Check dependencies:", dependency)
			checkDependencyAssetBundle(singleManifest, dependency, basePath, versionId)
		}
	}

	assetPath := filepath.FromSlash(basePath + string(os.PathSeparator) + name)

	// 新しいファイルor更新ファイル判断
	judgeNewOrUpdate(fileMap[name], versionId, name, assetPath, AssetBundleUploadStartURLPath, UpdateJudgeStrCRC, dependencies, FileInfo{})
}

func createResourcesUploadStartFileMap(basePath string, versionId int, recursion bool) {

	// get and make filepath list
	makeFilePathsForMulti(basePath, recursion, pathMap)

	// ファイル名で半角スペースがあるかチェック
	for _, filePath := range pathMap {
		checkFileNameWithSpace(filePath)
	}
	printErrorFile()

	for _, path := range pathMap {

		fileInfo, err := getFileInfo(path)
		if err != nil {
			log.Fatal(err)
			errorMsg := fmt.Sprintf("File '%v' have a error %v", path, err)
			ErrorFileMap[path] = errorMsg
		}

		if fileInfo.IsDir {
			continue
		}
		if len(fileInfo.Name) == 0 || fileInfo.Size == 0 {
			continue
		}

		serverFileData := fileMap[fileInfo.Name]
		//judgeNewOrUpdateResource(serverFileData, fileInfo, versionId, path, startUrlString)
		judgeNewOrUpdate(serverFileData, versionId, fileInfo.Name, path, ResourcesUploadStartURLPath, UpdateJudgeStrMD5, nil, fileInfo)

	}
	printErrorFile()

}

func getListOnServer(versionId int, listUrlStr string) error {
	listUrl := fmt.Sprintf(listUrlStr, Conf.Api.BaseUrl, versionId)
	err := utils.HttpGet(listUrl, &serverFileMap)
	if err != nil {
		utils.Fatal(err)
	}
	fileMap = serverFileMap["Files"].(map[string]interface{})
	return err
}

func createBucket(versionId int, cors bool, corsStr, fileType string) *GoogleCloudStorage {
	gcs := NewGoogleCloudStorage(serverFileMap["ProjectId"].(string), serverFileMap["Backet"].(string)+"-"+strconv.Itoa(versionId)+"-"+fileType, serverFileMap["Location"].(string))
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
	return gcs
}

func checkDir(basePath string) {
	fInfo, err := os.Stat(basePath)
	if err != nil {
		utils.Fatal(err)
	}
	if !fInfo.IsDir() {
		log.Fatal(basePath, "is not Directory.")
	}
}

func checkFile(filePath string) FileInfo {
	stat, err := os.Stat(filePath)
	if err != nil {
		utils.Fatal(err)
	}
	if stat.IsDir() {
		log.Fatal(filePath, "is Directory.")
	}
	fileInfo, err := getFileInfo(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if len(fileInfo.Name) == 0 || fileInfo.Size == 0 {
		log.Fatal(filePath, "is invalid.")
	}
	checkFileNameWithSpace(fileInfo.Name)

	printErrorFile()

	return fileInfo
}

// checkFileNameWithSpace ファイル名に半角スペースがあるかチェックしてある場合、エラーをmapに保存する
func checkFileNameWithSpace(fileName string) {
	space := isHaveSpace(fileName)
	if space {
		errorMsg := fmt.Sprintf("File '%v' have a space! please delete space on filename", fileName)
		ErrorFileMap[fileName] = errorMsg
	}
}

// printErrorFile ファイル名に半角がある場合、そのリストを出力して終了
func printErrorFile() {
	if len(ErrorFileMap) > 0 {
		for _, errMsg := range ErrorFileMap {
			log.Println(errMsg)
		}
		log.Fatal("have a problem that a file")
		os.Exit(1)
		return
	}
}

// isHaveSpace 半角スペースがあるかチェック
func isHaveSpace(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' {
			return true
		}
	}
	return false
}

func makeFilePathsForMulti(fpath string, isRecursion bool, pathMap map[string]string) {
	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		log.Println("makeFileInfos read dir error")
		utils.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() && isRecursion {
			makeFilePathsForMulti(filepath.Join(fpath, file.Name()), isRecursion, pathMap)
			continue
		}
		pathMap[file.Name()] = filepath.Join(fpath, file.Name())
	}
}

func startUploadMetaDataWithRetry(startUrl string, m *map[string]interface{}) error {
	err := try.Do(func(attempt int) (bool, error) {
		err := utils.HttpGet(startUrl, &m)
		if err != nil {
			time.Sleep(retryInterval)
			return attempt < 3, err
		}
		return false, nil
	})
	return err
}

func getFileInfo(path string) (FileInfo, error) {

	file, err := os.Open(path)
	if err != nil {
		file.Close()
		return FileInfo{0, "", "", false}, errors.Wrapf(err, "Error opening %q", path)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return FileInfo{0, "", "", false}, errors.Wrapf(err, "Error get stat %q", path)
	}

	if stat.IsDir() {
		file.Close()
		return FileInfo{0, "", "", true}, nil
	}
	size := stat.Size()
	md5 := getMD5Hash(file)
	file.Close()

	return FileInfo{size, md5, stat.Name(), false}, nil
}

func getMD5Hash(file *os.File) string {
	hasher := md5.New()
	_, err := io.Copy(hasher, file)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
