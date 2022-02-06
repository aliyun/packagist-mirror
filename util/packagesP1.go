package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (ctx *Context) SyncPackagesV1(processName string) {
	for {
		jobJson := sPop(packageP1Queue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		jobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &jobMap)
		if err != nil {
			fmt.Println(processName, "JSON Decode Error:", jobJson)
			sAdd(packageV1Set+"-json_decode_error", jobJson)
			continue
		}

		path, ok := jobMap["path"]
		if !ok {
			fmt.Println(processName, "package field not found: path")
			continue
		}

		hash, ok := jobMap["hash"]
		if !ok {
			fmt.Println(processName, "package field not found: hash")
			continue
		}

		key, ok := jobMap["key"]
		if !ok {
			fmt.Println(processName, "package field not found: key")
			continue
		}

		content, err := ctx.packagist.GetPackage(path)
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			makeFailed(packageV1Set, path, err)
			continue
		}

		// if resp.StatusCode != 200 {
		// 	syncHasError = true

		// 	// Make failed count
		// 	makeStatusCodeFailed(packageV1Set, resp.StatusCode, path)

		// 	// Push into failed queue to retry
		// 	if resp.StatusCode != 404 && resp.StatusCode != 410 {
		// 		sAdd(packageP1Queue, jobJson)
		// 	}

		// 	continue
		// }

		// content, err := ioutil.ReadAll(resp.Body)
		// _ = resp.Body.Close()
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			continue
		}

		// content, err = decode(content)
		// if err != nil {
		// 	syncHasError = true
		// 	fmt.Println("parseGzip Error", err.Error())
		// 	continue
		// }

		if sum := getSha256Sum(content); sum != hash {
			fmt.Println(processName, "Wrong Hash", "Original:", hash, "Current:", sum)
			syncHasError = true
			continue
		}

		// Put to OSS
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err = ctx.ossBucket.PutObject(path, bytes.NewReader(content), options...)
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

		hSet(packageV1Set, key, hash)
		dispatchDists(distMap["packages"], processName, ctx.mirror.distUrl+path)
		ctx.cdn.WarmUp(path)
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
