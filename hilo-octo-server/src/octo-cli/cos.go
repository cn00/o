package main

// 腾讯云
import (
	"fmt"
	"github.com/codegangsta/cli"
	//"github.com/tencentyun/cos-go-sdk-v5/debug"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type TencentCOS struct {
	//endpoint, accesKey, secret string
	client *cos.Client
}

func mainCos() {
	//将<bucketname>、<appid>和<region>修改为真实的信息
	//例如：http://test-1253846586.cos.ap-guangzhou.myqcloud.com
	u, _ := url.Parse("http://<bucketname>-<appid>.cos.<region>.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  os.Getenv("COS_SECRETID"),
			SecretKey: os.Getenv("COS_SECRETKEY"),
		},
	})
	//对象键（Key）是对象在存储桶中的唯一标识。
	//例如，在对象的访问域名 ` bucket1-1250000000.cos.ap-guangzhou.myqcloud.com/test/objectPut.go ` 中，对象键为 test/objectPut.go
	name := "test/objectPut.go"
	//Local file
	f := strings.NewReader("test")
	_, err := c.Object.Put(context.Background(), name, f, nil)
	if err != nil {
		panic(err)
	}
}

func (oss *TencentCOS) startUploadToOSS(priority int, tags cli.StringSlice, buildNumber, uploadType string, fileMap *FileMap) int {
	//var wg sync2.WaitGroup
	chanNum := runtime.NumCPU() * 100
	sem := make(chan string, chanNum)
	uploadChan := make(chan OSSFile, chanNum)
	count := len(fileMap.planUploadFileMap)

	log.Println("startUploadToGCP", uploadType, count, chanNum)

	//wg.Add(count)
	errorCount := 0
	idx := 0
	for k, v := range fileMap.planUploadFileMap {
		idx++
		go func(kk string, f GCSFile, idxi int, errC *int) {
			sem <- fmt.Sprintf("%d:%s", idxi, kk)
			//fileChan <- gcs.uploadWithChan(f.FilePath, f.FileName, f.AssetBundleName, f.ObjectName, f.Dependencies)
			//uploadErr := oss.uplaodWithRetry(f.FilePath, f.AssetBundleName)
			uploadErr := oss.uplaodWithRetry(f.FilePath, f.ObjectName)
			//uploadErr := error(nil)
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

	for range fileMap.planUploadFileMap {
		select {
		case f := <-uploadChan:
			log.Println("cosUpload", f.idx, "/", count, f.ObjectName, f.FilePath, f.Err)
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

func (tcos *TencentCOS) uplaodWithRetry(localPath, remotePath string) error {
	options := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: "application/octet-stream",
		},
		//ACLHeaderOptions: &cos.ACLHeaderOptions{
		//	XCosACL: "public",
		//},
	}
	if !strings.HasPrefix(remotePath, Conf.Cos.RootDir) {
		remotePath = Conf.Cos.RootDir + "/" + remotePath
	}

	doOnce := func() error {
		res, err := tcos.client.Object.PutFromFile(context.Background(), remotePath, localPath, options)
		if err != nil {
			log.Println("tcos.PutFromFile", remotePath, localPath, err, res.Status, res.Body)
		}
		return err
	}

	var err error
	for i := 0; i < 3; i++ {
		err = doOnce()
		if err == nil {
			return nil
		}
	}

	return err
}

func NewTencentCOS() *TencentCOS {
	var aloss = new(TencentCOS)
	u, _ := url.Parse(Conf.Cos.BaseUrl) //"http://<bucketname>-<appid>.cos.<region>.myqcloud.com"
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			//如实填写账号和密钥，也可以设置为环境变量
			SecretID:  Conf.Cos.SecretID,
			SecretKey: Conf.Cos.SecretKey,
			//Transport: &debug.DebugRequestTransport{
			//	RequestHeader: true,
			//	// Notice when put a large file and set need the request body, might happend out of memory error.
			//	RequestBody:    false,
			//	ResponseHeader: true,
			//	ResponseBody:   false,
			//},
		},
	})
	aloss.client = client

	log.Println("TencentCOSConfig", Conf.Cos.BaseUrl, Conf.Cos.SecretID, Conf.Cos.SecretKey)
	return aloss
}
