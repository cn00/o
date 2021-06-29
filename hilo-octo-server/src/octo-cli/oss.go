package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	_ "os"
	_"fmt"
	log"log"
	_"google.golang.org/api/classroom/v1"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	_ "github.com/aliyun/ossutil/lib"
	"runtime"
	"strings"
)

type AliyunOSS struct{
	//endpoint, accesKey, secret string
	client *oss.Client
	bucket *oss.Bucket
}

type OSSFile struct {
	FilePath        string
	ObjectName      string
	//MediaLink       string
	//Dependencies    []string
	Err             error
	idx				int
}

func (oss *AliyunOSS)startUploadToOSS(priority int, tags cli.StringSlice, buildNumber, uploadType string, fileMap *FileMap) int {
	//var wg sync2.WaitGroup
	chanNum := runtime.NumCPU()*10
	sem := make(chan string, chanNum)
	uploadChan := make(chan OSSFile, chanNum)
	count := len(fileMap.planUploadFileMap)

	log.Println("startUploadToGCP", uploadType, count, chanNum)

	//wg.Add(count)
	errorCount := 0
	idx := 0
	for k, v := range fileMap.planUploadFileMap {
		idx ++
		go func(kk string, f GCSFile, idxi int, errC *int) {
			sem <- fmt.Sprintf("%d:%s", idxi, kk)
			//fileChan <- gcs.uploadWithChan(f.FilePath, f.FileName, f.AssetBundleName, f.ObjectName, f.Dependencies)
			//uploadErr := oss.uplaodWithRetry(f.FilePath, f.AssetBundleName)
			uploadErr := oss.uplaodWithRetry(f.FilePath, f.ObjectName)
			defer func() {
				<-sem
				//<- uploadChan
				//wg.Done()
			}()
			gcsFile := GCSFile{f.FilePath, f.FileName, f.AssetBundleName,
				f.ObjectName, "", f.Dependencies, uploadErr}
			ossFile := OSSFile{f.FilePath, f.ObjectName, uploadErr, idxi}
			if uploadErr == nil {
				createUploadedNewFileList(gcsFile, priority, tags, buildNumber, uploadType, fileMap)
			} else {
				*errC += 1
			}
			uploadChan <- ossFile
		}(k, v, idx, &errorCount)
	}
	//wg.Wait()

	for range fileMap.planUploadFileMap{
		select {
		case f := <- uploadChan:
			log.Println("ossUpload", f.idx, "/", count, f.ObjectName, f.FilePath, f.Err)
		//case semi := <- sem:
		//	 log.Println("ossUpload", semi)
		}
	}

	defer close(sem)
	defer close(uploadChan)
	if errorCount > 0 {
		// 上传错误1个以上，中止
		log.Fatalln("Had error while uploading and cancel commit", errorCount)
	}
	return count
}

func (aloss *AliyunOSS)uplaodWithRetry(localPath, remotePath string) error{
	options := []oss.Option{
		oss.ContentType("application/octet-stream"),
	}
	if !strings.HasPrefix(remotePath, Conf.Oss.RootDir) {
		remotePath = Conf.Oss.RootDir + "/" + remotePath
	}

	doOnce := func() error {
		return aloss.bucket.PutObjectFromFile(remotePath, localPath, options...)
	}

	var  err error
	for i := 0; i < 3; i++ {
		err = doOnce()
		if err == nil{
			return nil
		}
	}
	
	return  err
}

func NewAliyunOSS() *AliyunOSS {
	var  aloss = new(AliyunOSS)
	client, err := oss.New(Conf.Oss.Endpoint, Conf.Oss.AccessKey, Conf.Oss.AccessSecret)
	if err != nil {
		log.Fatal(err)
	}
	aloss.client = client
	
	aloss.bucket, err = client.Bucket(Conf.Oss.Bucket) //Conf.Oss.Bucket
	if err != nil {
		log.Fatal(err)
	}
	return aloss
}
