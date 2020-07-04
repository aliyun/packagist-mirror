package util

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"time"
)

func ossClient(endpoint string) *oss.Client {
	var ossClient, err = oss.New(endpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if err != nil {
		errHandler(err)
	}
	return ossClient
}

func bucket(bucketName string, endpoint string) *oss.Bucket {
	bucket, err := ossClient(endpoint).Bucket(bucketName)
	if err != nil {
		errHandler(err)
	}
	return bucket
}

func putObject(processName string, objectKey string, reader io.Reader, options ...oss.Option) error {
	startT := time.Now()
	tc := time.Since(startT)

	err := bucket(config.OSSBucket, config.OSSEndpoint).PutObject(objectKey, reader, options...)
	if err != nil {
		fmt.Println(processName, "OSS Error", tc, mirrorUrl(objectKey), err.Error())
	} else {
		fmt.Println(processName, "OSS Put", tc, mirrorUrl(objectKey))
	}

	return err
}

func deleteObject(processName string, objectName string) {
	startT := time.Now()
	tc := time.Since(startT)

	err := bucket(config.OSSBucket, config.OSSEndpoint).DeleteObject(objectName)
	if err != nil {
		fmt.Println(processName, "OSS Error", tc, mirrorUrl(objectName), err.Error())
	} else {
		fmt.Println(processName, "OSS Deleted", tc, mirrorUrl(objectName))
	}
}

func isObjectExist(processName string, objectKey string) bool {
	isExist, err := bucket(config.OSSBucket, config.OSSEndpoint).IsObjectExist(objectKey)
	if err != nil {
		fmt.Println(processName, config.OSSBucket, objectKey, err.Error())
		errHandler(err)
	}
	if isExist {
		fmt.Println(processName, "OSS Exist", mirrorUrl(objectKey))
		return true
	}
	return false
}
