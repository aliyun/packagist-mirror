package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (ctx *Context) SyncProviders(processName string) {

	for {

		jobJson := sPop(providerQueue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		jobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &jobMap)
		if err != nil {
			sAdd(providerSet+"-json_decode_error", jobJson)
			continue
		}

		path, ok := jobMap["path"]
		if !ok {
			fmt.Println(processName, "provider field not found: path")
			continue
		}

		hash, ok := jobMap["hash"]
		if !ok {
			fmt.Println(processName, "provider field not found: hash")
			continue
		}

		key, ok := jobMap["key"]
		if !ok {
			fmt.Println(processName, "provider field not found: key")
			continue
		}

		content, err := ctx.packagist.GetAllPackages()
		if err != nil {
			syncHasError = true
			fmt.Println(processName, path, err.Error())
			makeFailed(providerSet, path, err)
			continue
		}

		// if resp.StatusCode != 200 {
		// 	syncHasError = true
		// 	// Make failed count
		// 	makeStatusCodeFailed(providerSet, resp.StatusCode, path)

		// 	// Push into failed queue to retry
		// 	if resp.StatusCode != 404 && resp.StatusCode != 410 {
		// 		pushProvider(key, path, hash, processName)
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

		if !CheckHash(processName, hash, content) {
			pushProvider(key, path, hash, processName)
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
			fmt.Println(processName, path, err.Error())
			errHandler(err)
			continue
		}

		hSet(providerSet, key, hash)

		dispatchPackages(distMap["providers"])

		ctx.cdn.WarmUp(path)
	}

}

func dispatchPackages(distMap interface{}) {
	for packageName, value := range distMap.(map[string]interface{}) {
		for _, hash := range value.(map[string]interface{}) {
			sha256 := hash.(string)
			if !hGetValue(packageV1Set, packageName, sha256) {
				p1 := make(map[string]interface{})
				p1["key"] = packageName
				p1["path"] = "p/" + packageName + "$" + sha256 + ".json"
				p1["hash"] = sha256
				jsonP1, _ := json.Marshal(p1)
				sAdd(packageP1Queue, string(jsonP1))
				countToday(packageV1Set, packageName)
			}
		}
	}
}
