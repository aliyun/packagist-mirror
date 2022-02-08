package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

func (ctx *Context) SyncPackagesV2(processName string) {
	var logger = NewLogger(processName)
	for {
		jobJson, err := ctx.redis.SPop(packageV2Queue).Result()
		if err == redis.Nil {
			logger.Info("get no task from " + providerQueue + ", sleep 1 second")
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			logger.Error("get task from " + packageV2Queue + " failed. " + err.Error())
			continue
		}

		// Json decode
		action, err := NewChangeActionFromJSONString(jobJson)
		if err != nil {
			logger.Error("unmarshal change action task failed. " + err.Error())
			continue
		}

		err = doAction(ctx, logger, action)
		if err != nil {
			logger.Error(fmt.Sprintf("process package action(%s). ", jobJson) + err.Error())
			continue
		}
	}

}

func doAction(ctx *Context, logger *MyLogger, action *ChangeAction) (err error) {

	actionType := action.Type

	if actionType == "update" {
		err = updatePackageV2(ctx, logger, action)
		return
	}

	if actionType == "delete" {
		err = deletePackageV2(ctx, logger, action)
		return
	}

	err = fmt.Errorf("unsupported action: %s", actionType)
	return
}

func updatePackageV2(ctx *Context, logger *MyLogger, action *ChangeAction) (err error) {
	packageName := action.Package
	content, err := ctx.packagist.GetPackage(packageName)
	if err != nil {
		// makeFailed(packageV2Set, packageName, err)
		return
	}

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
		return
	}

	_, err = ctx.redis.HSet(packageV2Set, packageName, action.Time).Result()
	if err != nil {
		return
	}

	ctx.cdn.WarmUp(path)
	return
}

func deletePackageV2(ctx *Context, logger *MyLogger, action *ChangeAction) (err error) {
	packageName := action.Package
	path := "p2/" + packageName + ".json"
	err = ctx.ossBucket.DeleteObject(path)
	if err != nil {
		return
	}

	_, err = ctx.redis.HDel(packageV2Set, packageName).Result()
	return
}
