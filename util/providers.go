package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

func providers(name string, num int) {

	processName := getProcessName(name, num)

	for {
		// BLPOP
		job, err := popFromQueue(providerHashFileKey)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		path := job[1]

		removeFromQueue(providerHashFileKey, path, processName)

		// Redis IsObjectExist
		if isSucceed(providerHashFileKey, path) {
			fmt.Println(processName, "Processed", path)
			continue
		}

		resp, err := packagistGet(path, processName)
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			makeFailed(providerHashFileKey, path, err.Error())
			continue
		}

		if resp.StatusCode != 200 {
			syncHasError = true

			// Make failed count
			makeStatusCodeFailed(providerHashFileKey, resp.StatusCode, path, packagistUrl(path))

			// Push into failed queue to retry
			if resp.StatusCode != 404 && resp.StatusCode != 410 {
				pushToQueueForStatusCodeRetry(providerHashFileKey, resp.StatusCode, path)
			}

			continue
		}

		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			continue
		}

		content, err = decode(content)
		if err != nil {
			syncHasError = true
			fmt.Println("parseGzip Error", err.Error())
			continue
		}

		if !CheckHash(processName, path, content) {
			pushProviderToQueue(path, processName)
			continue
		}

		// Put to OSS
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err = putObject(processName, path, bytes.NewReader(content), options...)
		if err != nil {
			syncHasError = true
			fmt.Println("putObject Error", err.Error())
			continue
		}

		// Json decode
		distMap := make(map[string]interface{})
		err = json.Unmarshal(content, &distMap)
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			errHandler(err)
			continue
		}

		dispatchPackages(distMap["providers"], getProcessName(name, num))

		// Mark succeed
		makeSucceed(providerHashFileKey, path, getProcessName(name, num))

		cdnCache(path, getProcessName(name, num))

	}

}

func dispatchPackages(distMap interface{}, processName string) {
	for packageName, value := range distMap.(map[string]interface{}) {

		for _, hash := range value.(map[string]interface{}) {

			path := "p/" + packageName + "$" + hash.(string) + ".json"

			if isSucceed(packageHashFileKey, path) {
				fmt.Println(processName, "Succeed", mirrorUrl(path))
				continue
			}

			pushPackageToQueue(path, processName)
			addIntoProcessing(path)
			count(packageKey, packageName)

		}

	}

}

func pushPackageToQueue(path string, processName string) {
	if pushToQueue(packageHashFileKey, path, processName) {

		count(packageHashFileKey, path)
	}
}
