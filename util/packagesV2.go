package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func (ctx *Context) SyncPackagesV2(processName string) {

	for {
		jobJson := sPop(packageV2Queue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		jobMap := make(map[string]string)
		err := json.Unmarshal([]byte(jobJson), &jobMap)
		if err != nil {
			fmt.Println(processName, "JSON Decode Error:", jobJson)
			sAdd(packageV2Set+"-json_decode_error", jobJson)
			continue
		}

		actionType, ok := jobMap["type"]
		if !ok {
			fmt.Println(processName, "package field not found: type")
			continue
		}

		if actionType == "update" {
			ctx.updatePackageV2(jobMap)
		}

		if actionType == "delete" {
			ctx.deletePackageV2(jobMap)
		}

	}

}

func (ctx *Context) updatePackageV2(jobMap map[string]string) {
	packageName, ok := jobMap["package"]
	if !ok {
		fmt.Println("package field not found: package")
		return
	}

	updateTime, ok := jobMap["time"]
	if !ok {
		fmt.Println("package field not found: time")
		return
	}

	content, err := ctx.packagist.GetPackage(packageName)
	if err != nil {
		makeFailed(packageV2Set, packageName, err)
		return
	}

	// if resp.StatusCode != 200 {
	// 	makeStatusCodeFailed(packageV2Set, resp.StatusCode, path)
	// 	return
	// }

	// JSON Decode
	packageJson := make(map[string]interface{})
	err = json.Unmarshal(content, &packageJson)
	if err != nil {
		fmt.Println("JSON Decode Error:", packageName)
		return
	}

	_, ok = packageJson["minified"]
	if !ok {
		fmt.Println("package field not found: minified")
		return
	}

	// Put to OSS
	options := []oss.Option{
		oss.ContentType("application/json"),
	}
	path := "p2/" + packageName + ".json"
	err = ctx.ossBucket.PutObject(path, bytes.NewReader(content), options...)
	if err != nil {
		syncHasError = true
		fmt.Println("putObject Error", err.Error())
		return
	}

	hSet(packageV2Set, packageName, updateTime)

	ctx.cdn.WarmUp(path)
}

func (ctx *Context) deletePackageV2(jobMap map[string]string) {
	packageName, ok := jobMap["package"]
	if !ok {
		fmt.Println("package field not found: package")
		return
	}

	path := "p2/" + packageName + ".json"

	hDel(packageV2Set, packageName)
	ctx.ossBucket.DeleteObject(path)
}
