package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

func packages(name string, num int) {

	processName := getProcessName(name, num)

	for {
		// BLPOP
		job, err := popFromQueue(packageHashFileKey)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		path := job[1]

		removeFromQueue(packageHashFileKey, path, processName)

		// Redis IsObjectExist
		if isSucceed(packageHashFileKey, path) {
			fmt.Println(processName, "Processed", path)
			continue
		}

		resp, err := packagistGet(path, processName)
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			makeFailed(packageHashFileKey, path, err.Error())
			continue
		}

		if resp.StatusCode != 200 {
			syncHasError = true

			// Make failed count
			makeStatusCodeFailed(packageHashFileKey, resp.StatusCode, path, packagistUrl(path))

			// Push into failed queue to retry
			if resp.StatusCode != 404 && resp.StatusCode != 410 {
				pushToQueueForStatusCodeRetry(packageHashFileKey, resp.StatusCode, path)
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
			pushPackageToQueue(path, processName)
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
			fmt.Println(processName, path, "Error parsing JSON", err.Error())
			continue
		}

		dispatchDists(distMap["packages"], processName, packagistUrl(path))

		makeSucceed(packageHashFileKey, path, processName)

		cdnCache(path, processName)

	}

}

func dispatchDists(packages interface{}, processName string, path string) {

	list, ok := packages.(map[string]interface{})
	if !ok {
		count(packagesNoData, path)
		return
	}

	for packageName, value := range list {
		for version, versionContent := range value.(map[string]interface{}) {

			distName := packageName + "/" + version

			for diskKeyName, dist := range versionContent.(map[string]interface{}) {

				if diskKeyName != "dist" {
					continue
				}

				distContent, ok := dist.(map[string]interface{})
				if !ok {
					redisClient.HSet(distsNoMetaKey, distName, path)
					continue
				}

				if v, ok := distContent["type"]; !ok {
					fmt.Println(processName, "type does not exist")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "type is empty")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				}

				if v, ok := distContent["url"]; !ok {
					fmt.Println(processName, "url does not exist")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "url is empty")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				}

				if v, ok := distContent["reference"]; !ok {
					fmt.Println(processName, "reference does not exist")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "reference is empty")
					redisClient.HSet(distsNoMetaKey, distName, path)

					continue
				}

				distJob := make(map[string]interface{})

				path := "dists/" + packageName + "/" + distContent["reference"].(string) + "." + distContent["type"].(string)

				distJob["path"] = path
				distJob["url"] = distContent["url"]
				distJob["reference"] = distContent["reference"]
				distJob["package"] = packageName
				distJob["version"] = version
				jsonString, _ := json.Marshal(distJob)

				if isSucceed(distsKey, path) {
					fmt.Println(processName, "Succeed", mirrorUrl(path))
					continue
				}

				pushDistToQueue(string(jsonString), path, distName, processName)
				addIntoProcessing(path)

			}

		}

	}

}

func pushDistToQueue(jsonString string, path string, distName string, processName string) {
	if pushToQueue(distsKey, jsonString, processName) {
		count(distsKey, path)
		count(versionsKey, distName)
	}
}
