package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

func packagesP2Dev(name string, num int) {

	for {
		jobJson := sPop(packageP2DevQueue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		JobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &JobMap)
		if err != nil {
			fmt.Println(getProcessName(name, num), "JSON Decode Error:", jobJson)
			sAdd(packageP2DevSet+"-json_decode_error", jobJson)
			continue
		}

		// Get information
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
			fmt.Println(getProcessName(name, num), path, err.Error())
			makeFailed(packageP2DevSet, path)
			continue
		}

		if resp.StatusCode != 200 {
			makeStatusCodeFailed(packageP2DevSet, resp.StatusCode, path)
			continue
		}

		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Println(getProcessName(name, num), path, err.Error())
			continue
		}

		content, err = decode(content)
		if err != nil {
			fmt.Println("parseGzip Error", err.Error())
			continue
		}

		// JSON Decode
		packageJson := make(map[string]interface{})
		err = json.Unmarshal(content, &packageJson)
		if err != nil {
			fmt.Println(getProcessName(name, num), "JSON Decode Error:", path)
			continue
		}

		_, ok = packageJson["minified"]
		if !ok {
			fmt.Println(getProcessName(name, num), "package field not found: minified")
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

		hSet(packageP2DevSet, key, hash)

		cdnCache(path, name, num)
	}

}
