package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"github.com/ahmetb/go-linq/v3"
	"github.com/codegangsta/cli"
	"runtime"
	"strconv"

	//linq "github.com/ahmetb/go-linq/v3"
	"github.com/pkg/errors"
	"gopkg.in/matryer/try.v1"
	"hilo-octo-proto/go/octo"
	"io"
	"io/ioutil"
	"log"
	"octo-cli/utils"
	sync2 "sync"

	"os"
	"path/filepath"
	//"strconv"
	"strings"
	"time"
)

type UploadOption struct {
	tag         string
	priority    int
	buildNumber string
}

type File struct {
	Id  int
	CRC uint32
}

// NewFile 新建文件
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

// FileInfo 文件信息
type FileInfo struct {
	Size  int64
	MD5   string
	Name  string
	IsDir bool
}

// GCSFile GCS中上传文件后，获取信息
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

	// UploadTypeResources Resources上传类型
	UploadTypeResources = "resources"
	// UploadTypeResources AssetBundle上传类型
	UploadTypeAssetBundle = "assetbundle"

	// UpdateJudgeStrCRC CRCはAssetBundle更新确认时使用
	UpdateJudgeStrCRC = "CRC"
	// UpdateJudgeStrMD5 MD5はResources更新确认时使用
	UpdateJudgeStrMD5 = "MD5"

	// AssetBundleUploadAllURLPath AssetBundle上传最后处理时使用的路径
	AssetBundleUploadAllURLPath = "%s/v1/upload/all/%d"
	// ResourcesUploadAllURLPath Resources上传最后处理时使用的路径
	ResourcesUploadAllURLPath = "%s/v1/resource/upload/all/%d"

	// AssetBundleListURLPath 已上传到服务器上AssetBundle获取信息的路径
	AssetBundleListURLPath = "%s/v1/upload/list/%d"
	// ResourcesListURLPath 已上传到服务器上Resources获取信息的路径
	ResourcesListURLPath = "%s/v1/resource/upload/list/%d"

	// AssetBundleUploadStartURLPath AssetBundle上传开始路径
	AssetBundleUploadStartURLPath = "%s/v1/upload/start?version=%d&filename=%s"
	// ResourcesUploadStartURLPath Resources上传开始路径
	ResourcesUploadStartURLPath = "%s/v1/resource/upload/start?version=%d&filename=%s"

	// MaxOutStanding fileUpload利用方法chan buffer的数量
	//MaxOutStanding = 100 //(runtime.NumCPU())
)

//var (
//	uploadedFileList    = []NewFile{}
//
//	objectNameMap       = map[string]string{}
//	planUploadFileMap   = map[string]GCSFile{}
//	serverFileMap       = map[string]interface{}{}
//	fileMap             = map[string]interface{}{}
//	pathMap             = map[string]string{}
//	ErrorFileMap        = map[string]string{}
//
//)

type ServerFile struct {
	id, app_id, version_id, revision_id, filename,
	object_name, url, size, crc, generation, md5, tag,
	assets, dependency, priority, state, build_number,
	upload_versionid, upd_datetime string
}

type FileMap struct {
	uploadedFileList []NewFile

	objectNameMap     map[string]string
	planUploadFileMap map[string]GCSFile
	serverFileMap     map[string]interface{}
	fileMap           map[string]interface{}
	pathMap           map[string]string
	ErrorFileMap      map[string]string
	option            UploadOption
}

// AssetBundle上传处理顺序
// 1. prepareUploadでGCS的信息，然后单击bucket创建
// 2. AssetBundle的确认
//    特定AssetBundle只确认注册
//    创建上传所需的信息（是新建还是更新）
//    OCTO服务器上的文件确认APIとStart API运行
// 3. 2的信息来源GCS上传到
// 4. 上传结束OCTOのALL API运行
// MultiUploadAssetBundle AssetBundle上传（并行处理）
func MultiUploadAssetBundle(versionId int, manifestPath string, tags cli.StringSlice, priority int, useOldTagFlg bool,
	buildNumber string, cors bool, corsStr, specificManifest string) {

	start := time.Now()

	fileMap := FileMap{}
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

	gcs, _ := prepareUpload(versionId, AssetBundleListURLPath, cors, corsStr, UploadTypeAssetBundle, &fileMap)

	singleManifest := DecodeSingleManifest(manifestPath)
	basePath := filepath.Dir(manifestPath)

	// 创建上传信息
	createAssetBundleUploadInfo(singleManifest, basePath, specificManifest, versionId, &fileMap)

	depsChangedCount := len(fileMap.uploadedFileList)

	// GCPのGCS上传到
	assetCount := startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeAssetBundle, &fileMap)

	// 上传结果OCTO API发送到
	endUpload(useOldTagFlg, versionId, AssetBundleUploadAllURLPath, &fileMap)

	elapsed := time.Since(start)
	log.Printf("Uploaded AssetBundle count: %d, Deps changed count: %d, Elapsed Time: %f Seconds\n", assetCount, depsChangedCount, elapsed.Seconds())

	printErrorFile()
}

// Resource上传处理顺序
// 1. prepareUploadでGCS的信息，然后单击bucket创建
// 2. 只上传特定文件时确认文件
//    上传多个文件时Directory的确认
//    创建上传所需的信息（是新建还是更新）
//    OCTO服务器上的文件确认APIとStart API运行
// 3. 2的信息来源GCS上传到
// 4. 上传结束OCTOのALL API运行
// MultiUploadResources Resources上传（并行处理）
func MultiUploadResources(versionId int, basePath string, tags cli.StringSlice, priority int, useOldTagFlg bool,
	buildNumber, corsStr string, cors, recursion bool, specificFilePath string) {

	if useOldTagFlg && len(tags) > 0 {
		log.Fatal("useOldTag and tags can not specified at the same time")
	}

	start := time.Now()

	fileMap := FileMap{}
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

	gcs, _ := prepareUpload(versionId, ResourcesListURLPath, cors, corsStr, UploadTypeResources, &fileMap)

	//check directory
	checkDir(basePath)

	createResourceUploadInfo(basePath, versionId, recursion, &fileMap)

	// 想单独给的情况
	if len(specificFilePath) > 0 {
		multiUploadOneResource(versionId, tags, priority, useOldTagFlg, buildNumber, nil, start, specificFilePath, &fileMap)
		return
	}

	resourcesCount := startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeResources, &fileMap)

	if len(fileMap.uploadedFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	endUpload(useOldTagFlg, versionId, ResourcesUploadAllURLPath, &fileMap)

	elapsed := time.Since(start)
	log.Printf("Upload Resources count: %d, Elapsed Time: %f Seconds\n", resourcesCount, elapsed.Seconds())

	printErrorFile()
}

func createAssetBundleUploadInfo(singleManifest SingleManifest, basePath string, specificManifest string, versionId int, fileMap *FileMap) {
	// 上传前start api，然后object map创建
	var uploadUrlVisitedMap = map[string]bool{}
	createAssetBundleUploadStartFileMapAndObjectMap(uploadUrlVisitedMap, singleManifest, basePath, specificManifest, versionId, fileMap)
	for manifest := range uploadUrlVisitedMap {
		createUploadAssetBundle(manifest, singleManifest, basePath, fileMap)
	}
}

func createResourceUploadInfo(basePath string, versionId int, recursion bool, fileMap *FileMap) {

	pathMap := createResourcesUploadStartFileMap(basePath, versionId, recursion, fileMap)

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
		serverFileData := fileMap.fileMap[fileInfo.Name]
		if serverFileData == nil {
			objectName := fileMap.objectNameMap[fileInfo.Name]
			setUploadNewFileMap(path, fileInfo.Name, objectName, nil, fileMap)
		} else {
			checkUpdateResourceFile(path, serverFileData, fileMap)
		}
	}
	jsonbyte, _ := json.Marshal(fileMap.planUploadFileMap)
	ioutil.WriteFile("planUploadFileMap.json", jsonbyte, os.ModePerm)
}

func prepareUpload(versionId int, listUrlString string, cors bool, corsStr, uploadType string, fileMap *FileMap) (*GoogleCloudStorage, error) {

	// 获取服务器上已有的文件列表
	err := getListOnServer(versionId, listUrlString, fileMap)
	if err != nil {
		utils.Fatal(err)
	}

	// gcs上bucket创建和cors設定
	//gcs := createBucket(versionId, cors, corsStr, uploadType, &fileMap)
	return nil, err
}

func multiUploadOneResource(versionId int, tags cli.StringSlice, priority int, useOldTagFlg bool, buildNumber string,
	gcs *GoogleCloudStorage, start time.Time, specificFilePath string, fileMap *FileMap) {

	//check file
	fileInfo := checkFile(specificFilePath)

	createOneResourceUploadInfo(fileInfo, versionId, specificFilePath, fileMap)

	resourcesCount := 0 // startUploadToGCP(gcs, priority, tags, buildNumber, UploadTypeResources)

	endUpload(useOldTagFlg, versionId, ResourcesUploadAllURLPath, fileMap)

	elapsed := time.Since(start)
	log.Printf("Upload Resources count: %d, Elapsed Time: %f Seconds\n", resourcesCount, elapsed.Seconds())

	printErrorFile()
}

func createOneResourceUploadInfo(fileInfo FileInfo, versionId int, specificFilePath string, fileMap *FileMap) {
	serverFileData := fileMap.fileMap[fileInfo.Name]
	judgeNewOrUpdate(serverFileData, versionId, fileInfo.Name, specificFilePath, ResourcesUploadStartURLPath, UpdateJudgeStrMD5, nil, fileInfo, fileMap)
	if serverFileData == nil {
		on := fileMap.objectNameMap[fileInfo.Name]
		setUploadNewFileMap(specificFilePath, fileInfo.Name, on, nil, fileMap)
	} else {
		checkUpdateResourceFile(specificFilePath, serverFileData, fileMap)
	}
}

func createUploadAssetBundle(manifest string, singleManifest SingleManifest, basePath string, fileMap *FileMap) {

	ds := singleManifest.AssetBundleManifest[manifest].Dependencies

	ap := filepath.FromSlash(basePath + "/" + manifest)
	data := fileMap.fileMap[manifest]
	if data == nil {
		on := fileMap.objectNameMap[manifest]
		setUploadNewFileMap(ap, manifest, on, ds, fileMap)
	} else {
		// 如果是已有文件，请检查该文件
		checkUpdateAssetBundleFile(ap, data, manifest, ds, fileMap)
	}
}

func checkUpdateAssetBundleFile(assetPath string, serverFileData interface{}, fileName string, dependency []string, fileMap *FileMap) {
	manifest := DecodeManifest(assetPath + ".manifest")
	crc := manifest.CRC
	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileCrc := uint32(serverFileDataM["CRC"].(float64))
	if serverFileCrc != crc || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("checkUpdateAssetBundleFile_CRC_up", encryptedName, fileName, serverFileCrc, "->", crc)
		setUploadNewFileMap(assetPath, fileName, encryptedName, dependency, fileMap)
	}
}

func checkUpdateResourceFile(assetPath string, serverFileData interface{}, fileMap *FileMap) {
	fileInfo, err := getFileInfo(assetPath)
	if err != nil {
	}
	serverFileDataM := serverFileData.(map[string]interface{})
	serverFileMD5 := serverFileDataM["MD5"].(string)
	if serverFileMD5 != fileInfo.MD5 || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
		encryptedName := serverFileDataM["EncriptedName"].(string)
		log.Println("checkUpdateResourceFile_MD5_up", encryptedName, fileInfo.Name, serverFileMD5, "->", fileInfo.MD5, assetPath)
		setUploadNewFileMap(assetPath, fileInfo.Name, encryptedName, nil, fileMap)
	}
}

func setUploadNewFileMap(filePath string, fileName string, objectName string, dependencies []string, fileMap *FileMap) {
	if len(objectName) > 0 {
		log.Println("setUploadNewFileMap", filePath, fileName, dependencies)
		fileMap.planUploadFileMap[fileName] = GCSFile{
			FilePath: filePath,
			FileName: fileName,
			// AssetBundle的情况下这边也设置
			AssetBundleName: fileName,
			ObjectName:      objectName,
			Dependencies:    dependencies,
		}
	}
}

func startUploadToGCP(gcs *GoogleCloudStorage, priority int, tags cli.StringSlice, buildNumber, uploadType string,
	fileMap *FileMap) int {
	var wg sync2.WaitGroup
	ncpu := runtime.NumCPU()
	sem := make(chan string, ncpu)
	fileChan := make(chan GCSFile, ncpu)
	count := len(fileMap.planUploadFileMap)

	log.Println("startUploadToGCP", uploadType, count)

	wg.Add(count)
	errorCount := 0
	for i, v := range fileMap.planUploadFileMap {
		go func(f GCSFile, ii string, errC *int) {
			sem <- ii
			defer func() { <-sem }()
			defer wg.Done()
			defer log.Println("<-sem:", ii)
			fileChan <- gcs.uploadWithChan(f.FilePath, f.FileName, f.AssetBundleName, f.ObjectName, f.Dependencies)
			gcsFile := GCSFile{f.FilePath, f.FileName, f.AssetBundleName, f.ObjectName, "", f.Dependencies, nil}
			if gcsFile.Err == nil {
				createUploadedNewFileList(gcsFile, priority, tags, buildNumber, uploadType, fileMap)
			} else {
				*errC += 1
				log.Printf("Upload error of %s: %+v errC:%d", gcsFile.AssetBundleName, gcsFile.Err, *errC)
			}
			fileChan <- gcsFile
		}(v, i, &errorCount)
	}

	for range fileMap.planUploadFileMap {
		//go func(f GCSFile, ii string) {
		<-fileChan
		//}(v, i)
	}
	wg.Wait()

	defer close(sem)
	if errorCount > 0 {
		// 上传错误1个以上，中止
		log.Fatalln("Had error while uploading and cancel commit")
	}
	return count

}

func createUploadedNewFileList(gcsFile GCSFile, priority int, tags cli.StringSlice, buildNumber, uploadType string, fileMap *FileMap) {
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
	serverFileData := fileMap.fileMap[gcsFile.FileName]
	if serverFileData == nil {
		if uploadType == UploadTypeAssetBundle {
			manifest := DecodeManifest(gcsFile.FilePath + ".manifest")
			crc := manifest.CRC
			nFile.Crc = crc
			nFile.Assets = manifest.Assets
		}
		fileMap.uploadedFileList = append(fileMap.uploadedFileList, nFile)
	} else {
		if uploadType == UploadTypeResources {
			serverFileDataM := serverFileData.(map[string]interface{})
			serverFileMD5 := serverFileDataM["MD5"].(string)
			if serverFileMD5 != nFile.MD5 || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
				nFile.EnciptName = serverFileDataM["EncriptedName"].(string)
				fileMap.uploadedFileList = append(fileMap.uploadedFileList, nFile)
			}
		} else if uploadType == UploadTypeAssetBundle {
			manifest := DecodeManifest(gcsFile.FilePath + ".manifest")
			crc := manifest.CRC
			nFile.Crc = crc
			nFile.Assets = manifest.Assets
			serverFileDataM := serverFileData.(map[string]interface{})
			serverFileCrc := uint32(serverFileDataM["CRC"].(float64))
			// TODO MD5添加比较
			if serverFileCrc != crc || octo.Data_State(serverFileDataM["State"].(float64)) == octo.Data_DELETE {
				encryptedName := serverFileDataM["EncriptedName"].(string)
				nFile.EnciptName = encryptedName
				fileMap.uploadedFileList = append(fileMap.uploadedFileList, nFile)
			}
		}

	}
}

func newFile(v2 GCSFile, priority int, tags cli.StringSlice, fileInfo FileInfo, crc uint32, buildNumber string) NewFile {

	newFile := createNewFile(v2, priority, tags, fileInfo, crc, buildNumber, fileInfo.Name)

	return newFile
}

func newFileAssetBundle(v2 GCSFile, priority int, tags cli.StringSlice, fileInfo FileInfo, crc uint32, buildNumber string, assetBundleName string) NewFile {
	// AsseBundleNameはPath的可能性，所以FileName利用方法
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

func endUpload(useOldTagFlg bool, versionId int, uploadUrlString string, fileMap *FileMap) {
	if len(fileMap.uploadedFileList) == 0 {
		log.Println("Nothing changed.")
		return
	}

	jsonBytes, err := json.Marshal(fileMap.uploadedFileList)
	if err != nil {
		panic(err)
	}

	if useOldTagFlg {
		uploadUrlString += "/notag"
	}
	uploadUrl := fmt.Sprintf(uploadUrlString, Conf.Api.BaseUrl, versionId)
	var res map[string]interface{}
	err = utils.HttpPost(uploadUrl, jsonBytes, &res)

	jsonBytesfmt, _ := json.MarshalIndent(fileMap.uploadedFileList, "", "  ")
	ioutil.WriteFile("log/uploadedFileList-"+fmt.Sprintf("%f", (res["RevisionId"].(float64)))+".json", jsonBytesfmt, 0666)
	if err != nil {
		log.Println("endUpload err: ", useOldTagFlg, versionId, uploadUrlString)
		utils.Fatal(err)
		return
	}
	log.Println("Upload Complete. RevisionId:", res["RevisionId"])
}

func createAssetBundleUploadStartFileMapAndObjectMap(uploadUrlVisitedMap map[string]bool, singleManifest SingleManifest,
	basePath, specificManifest string, versionId int, fileMap *FileMap) (SingleManifest, string) {

	// SingleManifest中的文件名是否有半角空间
	for name := range singleManifest.AssetBundleManifest {
		checkFileNameWithSpace(name)
	}

	// 如果文件名有半角，则输出该列表并退出
	printErrorFile()

	// SingleManifest中的一个Asset仅指定的情况
	if specificManifest != "" {
		if _, ok := singleManifest.AssetBundleManifest[specificManifest]; ok {
			uploadUrlVisitedMap[specificManifest] = true
			var dependencies = singleManifest.AssetBundleManifest[specificManifest].Dependencies
			assetPath := filepath.FromSlash(basePath + string(os.PathSeparator) + specificManifest)

			// 新建文件or确定更新文件
			judgeNewOrUpdate(fileMap.fileMap[specificManifest], versionId, specificManifest, assetPath,
				AssetBundleUploadStartURLPath, UpdateJudgeStrCRC, dependencies, FileInfo{}, fileMap)
		}
	} else {
		//var wg sync2.WaitGroup
		//ncpu := runtime.NumCPU()
		//sem := make(chan string, ncpu)
		//count := len(singleManifest.AssetBundleManifest)
		//log.Println("checkDependencyAssetBundle", count)
		//
		//wg.Add(count)

		// 如果不是指定特定文件，则递归检查依赖关系
		//idx := 0
		for name, _ := range singleManifest.AssetBundleManifest {
			//idx++
			//go func(nm string, i int) {
			//	sem <- nm
			//
			//	defer func() {
			//		<-sem
			//		wg.Done()
			//		log.Println("checkDependencyAssetBundle:", i, "/", count, nm)
			//	}()
			checkDependencyAssetBundle(uploadUrlVisitedMap, singleManifest, name, basePath, versionId, fileMap)
			//}(name, idx)
		}
		//wg.Wait()
		//defer close(sem)
	}
	return singleManifest, basePath
}

func judgeNewOrUpdate(serverFileData interface{}, versionId int, name, path, startUrlString, checkStr string,
	dependencies []string, fileInfo FileInfo, fileMap *FileMap) {

	if serverFileData == nil {
		// new
		log.Println("judgeNewOrUpdate_add_new", name, len(dependencies))

		var fileStartMap map[string]interface{}

		startURL := fmt.Sprintf(startUrlString, Conf.Api.BaseUrl, versionId, name)
		err := startUploadMetaDataWithRetry(startURL, &fileStartMap)
		if err != nil {
			log.Println("'%v' have a problem with err : %v", name, err)
			//ErrorFileMap[name] = errorMsg
		}
		fileMap.planUploadFileMap[name] = GCSFile{
			FilePath:     path,
			FileName:     name,
			ObjectName:   fileStartMap["FileName"].(string),
			Dependencies: dependencies,
		}
		fileMap.objectNameMap[name] = fileStartMap["FileName"].(string)
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
		log.Println("judgeNewOrUpdate_CRC_update", name)
		// had change in file
		fileMap.planUploadFileMap[name] = GCSFile{
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
		log.Println("judgeNewOrUpdate_dependencies_update", name)
		file := NewFile{
			Filename:    name,
			Dependency:  strings.Join(dependencies, ","),
			Tag:         fileMap.option.tag,
			Priority:    fileMap.option.priority,
			BuildNumber: fileMap.option.buildNumber,
		}
		fileMap.uploadedFileList = append(fileMap.uploadedFileList, file)
		return
	}
}

//var uploadUrlVisitedMapWLock = sync2.RWMutex{}
func checkDependencyAssetBundle(uploadUrlVisitedMap map[string]bool, singleManifest SingleManifest, name string,
	basePath string, versionId int, fileMap *FileMap) {
	var dependencies = singleManifest.AssetBundleManifest[name].Dependencies
	//if uploadUrlVisitedMap[name] {
	//	if false {
	//		log.Println("[DEBUG] already visited:", name)
	//	}
	//}

	//uploadUrlVisitedMapWLock.Lock()
	uploadUrlVisitedMap[name] = true

	// 检查依赖关系
	for _, dependency := range dependencies {
		if !uploadUrlVisitedMap[dependency] {
			//log.Println("Check dependencies:", name,  dependency)
			checkDependencyAssetBundle(uploadUrlVisitedMap, singleManifest, dependency, basePath, versionId, fileMap)
		}
	}
	//uploadUrlVisitedMapWLock.Unlock()

	assetPath := filepath.FromSlash(basePath + string(os.PathSeparator) + name)

	// 新建文件or确定更新文件
	judgeNewOrUpdate(fileMap.fileMap[name], versionId, name, assetPath, AssetBundleUploadStartURLPath, UpdateJudgeStrCRC, dependencies, FileInfo{}, fileMap)
}

func createResourcesUploadStartFileMap(basePath string, versionId int, recursion bool, fileMap *FileMap) map[string]string {

	// get and make filepath list
	pathMap := map[string]string{}
	makeFilePathsForMulti(basePath, recursion, pathMap)

	// 检查文件名是否有半角空间
	for _, filePath := range pathMap {
		checkFileNameWithSpace(filePath)
	}
	printErrorFile()

	for _, path := range pathMap {

		fileInfo, err := getFileInfo(path)
		if err != nil {
			//log.Fatal(err)
			log.Println("'%v' getFileInfo_error: %v", path, err)
			continue
			//ErrorFileMap[path] = errorMsg
		}

		if fileInfo.IsDir {
			continue
		}
		if len(fileInfo.Name) == 0 || fileInfo.Size == 0 {
			continue
		}

		serverFileData := fileMap.fileMap[fileInfo.Name]
		//judgeNewOrUpdateResource(serverFileData, fileInfo, versionId, path, startUrlString)
		judgeNewOrUpdate(serverFileData, versionId, fileInfo.Name, path, ResourcesUploadStartURLPath, UpdateJudgeStrMD5, nil, fileInfo, fileMap)

	}
	printErrorFile()
	return pathMap
}

func getListOnServer(versionId int, listUrlStr string, fileMap *FileMap) error {
	listUrl := fmt.Sprintf(listUrlStr, Conf.Api.BaseUrl, versionId)
	serverFileMap := map[string]interface{}{}
	err := utils.HttpGet(listUrl, &serverFileMap)
	if err != nil {
		utils.Fatal(err)
	}
	fs := serverFileMap["Files"]
	//fileMap := FileMap{}
	if fs != nil {
		fileMap.fileMap = fs.(map[string]interface{})
	} else {
		return fmt.Errorf("%s serverFileMap[\"Files\"] not exist", listUrlStr)
	}
	return err
}

func createBucket(versionId int, cors bool, corsStr, fileType string, serverFileMap map[string]interface{}) *GoogleCloudStorage {
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

// checkFileNameWithSpace 如果检查文件名是否有半角空间，请检查错误map保存到
func checkFileNameWithSpace(fileName string) {
	space := isHaveSpace(fileName)
	if space {
		errorMsg := fmt.Sprintf("'%v' have a space! please delete space on filename", fileName)
		//ErrorFileMap[fileName] = errorMsg
		log.Println("checkFileNameWithSpace", errorMsg)
	}
}

// printErrorFile 如果文件名有半角，则输出该列表并退出
func printErrorFile() {
	//if len(ErrorFileMap) > 0 {
	//	for _, errMsg := range ErrorFileMap {
	//		log.Println(errMsg)
	//	}
	//	log.Fatal("have a problem that a file")
	//	os.Exit(1)
	//	return
	//}
}

// isHaveSpace 检查是否有半角空间
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
	sha := hex.EncodeToString(hasher.Sum(nil))
	//log.Println("getMD5Hash", sha, file.Name())
	return sha
}
