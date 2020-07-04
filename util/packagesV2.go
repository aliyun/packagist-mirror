package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

func packagesV2(name string, num int) {

	for {
		jobJson := sPop(packageV2Queue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		JobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &JobMap)
		if err != nil {
			fmt.Println(getProcessName(name, num), "JSON Decode Error:", jobJson)
			sAdd(packageV2Set+"-json_decode_error", jobJson)
			continue
		}

		actionType, ok := JobMap["type"]
		if !ok {
			fmt.Println(getProcessName(name, num), "package field not found: type")
			continue
		}

		if actionType == "update" {
			updatePackageV2(JobMap, name, num)
		}

		if actionType == "delete" {
			deletePackageV2(JobMap, name, num)
		}

	}

}

func updatePackageV2(JobMap map[string]string, name string, num int) {
	packageName, ok := JobMap["package"]
	if !ok {
		fmt.Println(getProcessName(name, num), "package field not found: package")
		return
	}

	updateTime, ok := JobMap["time"]
	if !ok {
		fmt.Println(getProcessName(name, num), "package field not found: time")
		return
	}

	path := "p2/" + packageName + ".json"
	resp, err := packagistGet(path, getProcessName(name, num))
	if err != nil {
		fmt.Println(getProcessName(name, num), path, err.Error())
		makeFailed(packageV2Set, path, err)
		return
	}

	if resp.StatusCode != 200 {
		makeStatusCodeFailed(packageV2Set, resp.StatusCode, path)
		return
	}

	content, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Println(getProcessName(name, num), path, err.Error())
		return
	}

	content, err = decode(content)
	if err != nil {
		fmt.Println("parseGzip Error", err.Error())
		return
	}

	// JSON Decode
	packageJson := make(map[string]interface{})
	err = json.Unmarshal(content, &packageJson)
	if err != nil {
		fmt.Println(getProcessName(name, num), "JSON Decode Error:", path)
		return
	}

	_, ok = packageJson["minified"]
	if !ok {
		fmt.Println(getProcessName(name, num), "package field not found: minified")
		return
	}

	// Put to OSS
	options := []oss.Option{
		oss.ContentType("application/json"),
	}
	err = putObject(getProcessName(name, num), path, bytes.NewReader(content), options...)
	if err != nil {
		syncHasError = true
		fmt.Println("putObject Error", err.Error())
		return
	}

	hSet(packageV2Set, packageName, updateTime)

	cdnCache(path, name, num)
}

func deletePackageV2(JobMap map[string]string, name string, num int) {
	packageName, ok := JobMap["package"]
	if !ok {
		fmt.Println(getProcessName(name, num), "package field not found: package")
		return
	}

	path := "p2/" + packageName + ".json"

	hDel(packageV2Set, packageName)
	deleteObject(getProcessName(name, num), path)
}
