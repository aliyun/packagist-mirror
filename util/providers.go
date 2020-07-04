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
			fmt.Println(getProcessName(name, num), "provider field not found: path")
			continue
		}

		hash, ok := jobMap["hash"]
		if !ok {
			fmt.Println(getProcessName(name, num), "provider field not found: hash")
			continue
		}

		key, ok := jobMap["key"]
		if !ok {
			fmt.Println(getProcessName(name, num), "provider field not found: key")
			continue
		}

		resp, err := packagistGet(path, getProcessName(name, num))
		if err != nil {
			syncHasError = true
			fmt.Println(getProcessName(name, num), path, err.Error())
			makeFailed(providerSet, path, err)
			continue
		}

		if resp.StatusCode != 200 {
			syncHasError = true
			// Make failed count
			makeStatusCodeFailed(providerSet, resp.StatusCode, path)

			// Push into failed queue to retry
			if resp.StatusCode != 404 && resp.StatusCode != 410 {
				pushProvider(key, path, hash, getProcessName(name, num))
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
			pushProvider(key, path, hash, getProcessName(name, num))
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
			fmt.Println(getProcessName(name, num), path, err.Error())
			errHandler(err)
			continue
		}

		hSet(providerSet, key, hash)

		dispatchPackages(distMap["providers"])

		cdnCache(path, name, num)
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
