package main

import (
	_ "os"
	ossutil "github.com/aliyun/ossutil/lib"
)

func uplaod(endpoint, accessID, accessKey, localPath, remotePath string){
	//bucket, err := oss.New(endpoint, accessID, accessKey)
	//if err != nil {
	//	println(err)
	//}
	//bucket.
	args :=  []string{"cp", localPath, remotePath}
	ossutil.RunCommand(args, nil)
}