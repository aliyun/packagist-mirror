package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

func (ctx *Context) SyncProvider(processName string) {
	var logger = NewLogger(processName)
	for {
		jobJson, err := ctx.redis.SPop(providerQueue).Result()
		if err == redis.Nil {
			// logger.Info("get no task from " + providerQueue + ", sleep 1 second")
			time.Sleep(1 * time.Second)
			continue
		}

		if err != nil {
			logger.Error("get task from " + providerQueue + " failed. " + err.Error())
			continue
		}

		providerTask, err := NewTaskFromJSONString(jobJson)
		if err != nil {
			logger.Error("unmarshal provider task failed. " + err.Error())
			continue
		}

		logger.Info(fmt.Sprintf("dispatch provider: %s", providerTask.Key))
		err = syncProvider(ctx, logger, providerTask)
		if err != nil {
			logger.Error("sync provider failed. " + err.Error())
			continue
		}
	}

}

func syncProvider(ctx *Context, logger *MyLogger, task *Task) (err error) {
	content, err := ctx.packagist.Get(task.Path)
	if err != nil {
		return
	}

	if sum := getSha256Sum(content); sum != task.Hash {
		logger.Error("Wrong Hash, Original: " + task.Hash + " Current: " + sum)
		logger.Error("re-add into provider queue")
		jsonP2, _ := json.Marshal(task)
		_, err = ctx.redis.SAdd(providerQueue, string(jsonP2)).Result()
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

	err = ctx.redis.HSet(providerSet, task.Key, task.Hash).Err()
	if err != nil {
		return
	}

	// Json decode
	providersRoot, err := NewProvidersFromJSONString(string(content))
	if err != nil {
		return
	}

	for packageName, hashers := range providersRoot.Providers {
		sha256 := hashers.SHA256
		exists, err2 := ctx.redis.HExists(packageV1Set, packageName).Result()
		if err2 != nil {
			return
		}

		if exists {
			value, err2 := ctx.redis.HGet(packageV1Set, packageName).Result()
			if err2 != nil {
				return
			}

			if sha256 == value {
				continue
			}

			logger.Info(fmt.Sprintf("dispatch package(%s) to %s", packageName, packageP1Queue))
			task := NewTask(packageName, "p/"+packageName+"$"+sha256+".json", sha256)
			jsonP1, _ := json.Marshal(task)
			ctx.redis.SAdd(packageP1Queue, string(jsonP1)).Result()
			ctx.redis.SAdd(getTodayKey(packageV1Set), packageName)
			ctx.redis.ExpireAt(getTodayKey(packageV1Set), getTomorrow())
		}
	}

	ctx.cdn.WarmUp(task.Path)
	return
}
