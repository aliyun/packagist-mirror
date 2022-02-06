package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var packagistLastModified = ""
var syncHasError = false

func (ctx *Context) SyncPackagesJsonFile(processName string) {
	logger := log.New(os.Stderr, processName, log.LUTC)

	for {
		err := syncPackagesJsonFile(ctx, logger)
		if err != nil {
			logger.Println("sync packages.json failed: " + err.Error())
			return
		}
		// Each cycle requires a time slot
		time.Sleep(1 * time.Second)
	}

}

func getPackages(content []byte) (packages *Packages, err error) {
	packages = new(Packages)
	// JSON Decode
	err = json.Unmarshal(content, &packages)
	return
}

func syncPackagesJsonFile(ctx *Context, logger *log.Logger) (err error) {
	// Get root file from packagist repo
	packagistContent, err := ctx.packagist.GetPackagesJSON()

	if err != nil {
		err = fmt.Errorf("get packages.json failed: " + err.Error())
		return
	}

	localPackagesJsonSum, err := ctx.redis.Get(packagesJsonKey).Result()
	if err != nil {
		err = fmt.Errorf("get local packages.json sum failed: " + err.Error())
		return
	}

	sum := getSha256Sum(packagistContent)
	if localPackagesJsonSum == sum {
		// packages.json is not changed
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
		if err2 != nil {
			err = fmt.Errorf("get provider set with key(%s) failed: "+err2.Error(), provider)
			return
		}

		if value != providerHash {
			p := make(map[string]interface{})
			p["key"] = provider
			p["path"] = providerPath
			p["hash"] = providerHash
			jsonP2, _ := json.Marshal(p)
			ctx.redis.SAdd(providerQueue, string(jsonP2)).Result()
			ctx.redis.SAdd(getTodayKey(providerSet), providerHash).Result()
			ctx.redis.ExpireAt(getTodayKey(providerSet), getTomorrow()).Result()
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
		fmt.Println("Processing:", left, ", Check again in 1 second. ")
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
