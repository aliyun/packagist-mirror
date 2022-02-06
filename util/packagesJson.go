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

var packagesContentCache []byte
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

	if bytes.Equal(packagesContentCache, packagistContent) {
		return
	}

	// JSON Decode
	packagesJson, err := getPackages(packagistContent)
	if err != nil {
		err = fmt.Errorf("unmarshal packages.json failed: " + err.Error())
		return
	}

	// Make error false
	syncHasError = false

	// Dispatch providers
	for provider, hashValue := range packagesJson.ProviderIncludes {
		providerHash := hashValue.SHA256
		providerPath := strings.Replace(provider, "%hash%", providerHash, -1)

		if !hGetValue(providerSet, provider, providerHash) {
			p := make(map[string]interface{})
			p["key"] = provider
			p["path"] = providerPath
			p["hash"] = providerHash
			jsonP2, _ := json.Marshal(p)
			ctx.redis.SAdd(providerQueue, string(jsonP2)).Result()
			countToday(providerSet, providerHash)
		} else {
			fmt.Println("Already succeed")
		}
	}

	for {
		// If all tasks are completed, skip the loop and update the file
		left := sCard(distQueue) + sCard(providerQueue) + sCard(packageP1Queue) + sCard(packageV2Queue)
		if left == 0 {
			break
		}
		fmt.Println("Processing:", left, ", Check again in 1 second. ")
		time.Sleep(1 * time.Second)
	}

	if syncHasError == true {
		err = fmt.Errorf("There is an error in this synchronization. We look forward to the next synchronization...")
		return
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

	// The cache is updated only if the push is successful
	packagesContentCache = packagistContent
	return
}
