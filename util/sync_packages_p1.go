package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

type Package struct {
	Name string      `json:"name"`
	Dist PackageDist `json:"dist"`
}

type PackageDist struct {
	Type      string `json:"type"`
	Url       string `json:"url"`
	Reference string `json:"reference"`
	Shasum    string `json:"shasum"`
	// "type": "zip",
	// "url": "https://gitlab.com/api/v4/projects/ACP3%2Fcms/repository/archive.zip?sha=78b68105237832ec6684299f17857c58fa895a46",
	// "reference": "78b68105237832ec6684299f17857c58fa895a46",
	// "shasum": ""
}

type Response struct {
	Packages map[string]map[string]Package
}

func (ctx *Context) SyncPackagesV1(processName string) {
	var logger = NewLogger(processName)
	for {
		jobJson, err := ctx.redis.SPop(packageP1Queue).Result()
		if err == redis.Nil {
			// logger.Info("get no task from " + packageP1Queue + ", sleep 1 second")
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			logger.Error("get task from " + packageP1Queue + " failed. " + err.Error())
			continue
		}

		// Json decode
		task, err := NewTaskFromJSONString(jobJson)
		if err != nil {
			logger.Error("unmarshal package task failed. " + err.Error())
			continue
		}

		err = syncPackage(ctx, logger, task)
		if err != nil {
			logger.Error("sync package failed. " + err.Error())
			continue
		}
	}

}

func syncPackage(ctx *Context, logger *MyLogger, task *Task) (err error) {
	content, err := ctx.packagist.Get(task.Path)
	if err != nil {
		return
	}

	if sum := getSha256Sum(content); sum != task.Hash {
		logger.Error(fmt.Sprintf("Wrong Hash, Original: %s, Current: %s", task.Hash, sum))
		return
	}

	// Put to OSS
	options := []oss.Option{
		oss.ContentType("application/json"),
	}
	err = ctx.ossBucket.PutObject(task.Path, bytes.NewReader(content), options...)
	if err != nil {
		return
	}

	// Json decode
	response := new(Response)
	err = json.Unmarshal(content, &response)
	if err != nil {
		return
	}

	ctx.redis.HSet(packageV1Set, task.Key, task.Hash).Err()
	for packageName, versions := range response.Packages {
		for versionName, packageVersion := range versions {
			distName := packageName + "/" + versionName

			dist := packageVersion.Dist
			path := "dists/" + packageName + "/" + dist.Reference + "." + dist.Type

			exist, err2 := ctx.redis.SIsMember(distSet, path).Result()
			if err2 != nil {
				err = fmt.Errorf("check dists path failed: " + err2.Error())
				return
			}

			if !exist {
				dist := NewDist(path, dist.Url)
				ctx.redis.SAdd(distQueue, dist.ToJSONString())
				ctx.redis.SAdd(versionsSet, distName).Result()
				ctx.redis.SAdd(getTodayKey(versionsSet), distName)
				ctx.redis.ExpireAt(getTodayKey(versionsSet), getTomorrow())
			}
		}
	}

	ctx.cdn.WarmUp(task.Path)
	ctx.redis.SAdd(getTodayKey(packageV1SetHash), task.Path)
	ctx.redis.ExpireAt(getTodayKey(packageV1SetHash), getTomorrow())
	return
}
