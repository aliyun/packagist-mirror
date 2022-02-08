package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

func (ctx *Context) SyncPackagesJsonFile(processName string) {
	logger := NewLogger(processName)

	for {
		err := syncPackagesJsonFile(ctx, logger)
		if err != nil {
			logger.Error("sync packages.json failed: " + err.Error())
			return
		}
		// Each cycle requires a time slot
		time.Sleep(60 * time.Second)
	}

}

func getPackages(content []byte) (packages *Packages, err error) {
	packages = new(Packages)
	// JSON Decode
	err = json.Unmarshal(content, &packages)
	return
}

func syncPackagesJsonFile(ctx *Context, logger *MyLogger) (err error) {
	// Get root file from packagist repo
	logger.Info("get packages.json now")
	packagistContent, packagistLastModified, err := ctx.packagist.GetPackagesJSON()
	logger.Info("get packages.json done")
	if err != nil {
		err = fmt.Errorf("get packages.json failed: " + err.Error())
		return
	}

	err = ctx.redis.Set(packagistLastModifiedKey, packagistLastModified, 0).Err()
	if err != nil {
		return
	}

	sum := getSha256Sum(packagistContent)

	localPackagesJsonSum, err2 := ctx.redis.Get(packagesJsonKey).Result()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("get local packages.json sum failed: " + err2.Error())
		return
	}

	if localPackagesJsonSum == sum {
		// packages.json is not changed
		logger.Info("packages.json is not changed")
		return
	}

	// JSON Decode
	packagesJson, err := getPackages(packagistContent)
	if err != nil {
		err = fmt.Errorf("unmarshal packages.json failed: " + err.Error())
		return
	}

	// Dispatch providers
	for provider, hashValue := range packagesJson.ProviderIncludes {
		providerHash := hashValue.SHA256
		providerPath := strings.Replace(provider, "%hash%", providerHash, -1)

		value, err2 := ctx.redis.HGet(providerSet, provider).Result()
		if err2 == redis.Nil {
			logger.Info("dispatch providers: " + provider)
			task := NewTask(provider, providerPath, providerHash)
			jsonP2, _ := json.Marshal(task)
			ctx.redis.SAdd(providerQueue, string(jsonP2)).Result()
			ctx.redis.SAdd(getTodayKey(providerSet), providerHash).Result()
			ctx.redis.ExpireAt(getTodayKey(providerSet), getTomorrow()).Result()
			continue
		}

		if err2 != nil {
			err = fmt.Errorf("get provider set with key(%s) failed: "+err2.Error(), provider)
			return
		}

		if value != providerHash {
			logger.Info("dispatch providers: " + provider)
			task := NewTask(provider, providerPath, providerHash)
			jsonP2, _ := json.Marshal(task)
			ctx.redis.SAdd(providerQueue, string(jsonP2)).Result()
			ctx.redis.SAdd(getTodayKey(providerSet), providerHash).Result()
			ctx.redis.ExpireAt(getTodayKey(providerSet), getTomorrow()).Result()
			continue
		}
	}

	for {
		// If all tasks are completed, skip the loop and update the file
		distQueueSize, err1 := ctx.redis.SCard(distQueue).Result()
		if err1 != nil {
			err = fmt.Errorf("get distQueue size: " + err1.Error())
			return
		}

		providerQueueSize, err1 := ctx.redis.SCard(providerQueue).Result()
		if err1 != nil {
			err = fmt.Errorf("get providerQueue size: " + err1.Error())
			return
		}

		packageP1QueueSize, err1 := ctx.redis.SCard(packageP1Queue).Result()
		if err1 != nil {
			err = fmt.Errorf("get packageP1Queue size: " + err1.Error())
			return
		}

		packageV2QueueSize, err1 := ctx.redis.SCard(packageV2Queue).Result()
		if err1 != nil {
			err = fmt.Errorf("get packageV2Queue size: " + err1.Error())
			return
		}

		left := distQueueSize + providerQueueSize + packageP1QueueSize + packageV2QueueSize
		if left == 0 {
			break
		}
		logger.Info(fmt.Sprintf("Processing: %d, Check again in 1 second.", left))
		time.Sleep(1 * time.Second)
	}

	// Update `packages.json`
	var newPackagesJson = make(map[string]interface{})
	json.Unmarshal(packagistContent, &newPackagesJson)
	var lastUpdateTime = time.Now().Format("2006-01-02 15:04:05")
	// set to redis
	err = ctx.redis.Set(lastUpdateTimeKey, lastUpdateTime, 0).Err()
	if err != nil {
		return
	}

	newPackagesJson["last-update"] = lastUpdateTime
	newPackagesJson["metadata-url"] = ctx.mirror.providerUrl + "p2/%package%.json"
	newPackagesJson["providers-url"] = ctx.mirror.providerUrl + "p/%package%$%hash%.json"
	newPackagesJson["mirrors"] = []map[string]interface{}{
		{
			"dist-url":  ctx.mirror.distUrl + "dists/%package%/%reference%.%type%",
			"preferred": true,
		},
	}

	// Json Encode
	packagesJsonNew, _ := json.Marshal(newPackagesJson)

	// Update packages.json
	options := []oss.Option{
		oss.ContentType("application/json"),
	}
	err = ctx.ossBucket.PutObject("packages.json", bytes.NewReader(packagesJsonNew), options...)
	if err != nil {
		return
	}

	// save local packages.json sum to redis
	err = ctx.redis.Set(packagesJsonKey, sum, 0).Err()
	if err != nil {
		err = fmt.Errorf("save local packages.json sum failed: " + err.Error())
		return
	}

	return
}
