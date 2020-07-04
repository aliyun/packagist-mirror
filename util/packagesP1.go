package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

func packagesV1(name string, num int) {

	for {
		jobJson := sPop(packageP1Queue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		JobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &JobMap)
		if err != nil {
			fmt.Println(getProcessName(name, num), "JSON Decode Error:", jobJson)
			sAdd(packageV1Set+"-json_decode_error", jobJson)
			continue
		}

		path, ok := JobMap["path"]
		if !ok {
			fmt.Println(getProcessName(name, num), "package field not found: path")
			continue
		}

		hash, ok := JobMap["hash"]
		if !ok {
			fmt.Println(getProcessName(name, num), "package field not found: hash")
			continue
		}

		key, ok := JobMap["key"]
		if !ok {
			fmt.Println(getProcessName(name, num), "package field not found: key")
			continue
		}

		resp, err := packagistGet(path, getProcessName(name, num))
		if err != nil {
			syncHasError = true
			fmt.Println(getProcessName(name, num), path, err.Error())
			makeFailed(packageV1Set, path, err)
			continue
		}

		if resp.StatusCode != 200 {
			syncHasError = true

			// Make failed count
			makeStatusCodeFailed(packageV1Set, resp.StatusCode, path)

			// Push into failed queue to retry
			if resp.StatusCode != 404 && resp.StatusCode != 410 {
				sAdd(packageP1Queue, jobJson)
			}

			continue
		}

		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			syncHasError = true
			fmt.Println(getProcessName(name, num), path, err.Error())
			continue
		}

		content, err = decode(content)
		if err != nil {
			syncHasError = true
			fmt.Println("parseGzip Error", err.Error())
			continue
		}

		if !CheckHash(getProcessName(name, num), hash, content) {
			syncHasError = true
			continue
		}

		// Put to OSS
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err = putObject(getProcessName(name, num), path, bytes.NewReader(content), options...)
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
			fmt.Println(getProcessName(name, num), path, "Error parsing JSON", err.Error())
			continue
		}

		hSet(packageV1Set, key, hash)
		dispatchDists(distMap["packages"], getProcessName(name, num), packagistUrl(path))
		cdnCache(path, name, num)
		countToday(packageV1SetHash, path)
	}

}

func dispatchDists(packages interface{}, processName string, path string) {

	list, ok := packages.(map[string]interface{})
	if !ok {
		countAll(packagesNoData, path)
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
					sAdd(distsNoMetaKey, path)
					continue
				}

				if v, ok := distContent["type"]; !ok {
					fmt.Println(processName, "type does not exist")
					sAdd(distsNoMetaKey, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "type is empty")
					sAdd(distsNoMetaKey, path)

					continue
				}

				if v, ok := distContent["url"]; !ok {
					fmt.Println(processName, "url does not exist")
					sAdd(distsNoMetaKey, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "url is empty")
					sAdd(distsNoMetaKey, path)

					continue
				}

				if v, ok := distContent["reference"]; !ok {
					fmt.Println(processName, "reference does not exist")
					sAdd(distsNoMetaKey, path)

					continue
				} else if v == nil {
					fmt.Println(processName, "reference is empty")
					sAdd(distsNoMetaKey, path)

					continue
				}

				path := "dists/" + packageName + "/" + distContent["reference"].(string) + "." + distContent["type"].(string)

				if !sIsMember(distSet, path) {
					distJob := make(map[string]interface{})
					distJob["path"] = path
					distJob["url"] = distContent["url"]
					jsonString, _ := json.Marshal(distJob)
					sAdd(distQueue, string(jsonString))
					countAll(versionsSet, distName)
					countToday(versionsSet, distName)
				}

			}

		}

	}

}
