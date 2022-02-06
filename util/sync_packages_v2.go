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
		jobJson, err := ctx.redis.SPop(packageV2Queue).Result()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// Json decode
		action, err := NewChangeActionFromJSONString(jobJson)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		actionType := action.Type

		if actionType == "update" {
			updatePackageV2(ctx, action)
		}

		if actionType == "delete" {
			deletePackageV2(ctx, action)
		}

	}

}

func updatePackageV2(ctx *Context, action *ChangeAction) (err error) {
	packageName := action.Package
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

	_, ok := packageJson["minified"]
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

	_, err = ctx.redis.HSet(packageV2Set, packageName, action.Time).Result()
	if err != nil {
		return
	}

	ctx.cdn.WarmUp(path)
	return
}

func deletePackageV2(ctx *Context, action *ChangeAction) (err error) {
	packageName := action.Package
	path := "p2/" + packageName + ".json"
	err = ctx.ossBucket.DeleteObject(path)
	if err != nil {
		return
	}

	_, err = ctx.redis.HDel(packageV2Set, packageName).Result()
	return
}
