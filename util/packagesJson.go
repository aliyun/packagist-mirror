package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var packagesJson = make(map[string]interface{})
var packagesContentCache []byte
var packagistLastModified = ""
var syncHasError = false

func (ctx *Context) SyncPackagesJsonFile(processName string) {

	for {
		// Each cycle requires a time slot
		time.Sleep(1 * time.Second)

		// Get root file from packagist repo
		packagistContent, err := ctx.packagist.GetPackagesJSON()
		if err != nil {
			sAdd(packagesJsonKey+"-Get", "packages.json")
			continue
		}

		// Status code must be 200
		// if resp.StatusCode != 200 {
		// 	makeStatusCodeFailed(packagesJsonKey, resp.StatusCode, config.GetPackagistUrl("packages.json"))
		// 	continue
		// }

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// Decode content if Gzip
		packagistContent, err = decode(packagistContent)
		if err != nil {
			fmt.Println("parseGzip Error", err.Error())
			continue
		}

		if bytes.Equal(packagesContentCache, packagistContent) {
			continue
		}

		// JSON Decode
		err = json.Unmarshal(packagistContent, &packagesJson)
		if err != nil {
			sAdd("root-json_decode_error", "root")
			continue
		}

		// Make error false
		syncHasError = false

		// Dispatch providers
		ctx.dispatchProviders(packagesJson["provider-includes"], processName)

		for {
			// If all tasks are completed, skip the loop and update the file
			left := sCard(distQueue) + sCard(providerQueue) + sCard(packageP1Queue) + sCard(packageV2Queue)
			if left == 0 {
				break
			}
			fmt.Println(processName, "Processing:", left, ", Check again in 1 second. ")
			time.Sleep(1 * time.Second)
		}

		if syncHasError == true {
			fmt.Println(processName, "There is an error in this synchronization. We look forward to the next synchronization...")
			continue
		}

		// Update `packages.json`
		packagesJson["last-update"] = time.Now().Format("2006-01-02 15:04:05")
		packagesJson["metadata-url"] = ctx.mirror.providerUrl + "p2/%package%.json"
		packagesJson["providers-url"] = ctx.mirror.providerUrl + "p/%package%$%hash%.json"
		packagesJson["mirrors"] = []map[string]interface{}{
			{
				"dist-url":  ctx.mirror.distUrl + "dists/%package%/%reference%.%type%",
				"preferred": true,
			},
		}

		// Json Encode
		packagesJsonNew, _ := json.Marshal(packagesJson)

		// Update packages.json
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err = ctx.ossBucket.PutObject("packages.json", bytes.NewReader(packagesJsonNew), options...)
		if err != nil {
			continue
		}

		// The cache is updated only if the push is successful
		packagesContentCache = packagistContent
	}

}

func (ctx *Context) dispatchProviders(distMap interface{}, name string) {

	for provider, value := range distMap.(map[string]interface{}) {

		for _, hash := range value.(map[string]interface{}) {

			providerHash := hash.(string)
			providerPath := strings.Replace(provider, "%hash%", providerHash, -1)

			if !hGetValue(providerSet, provider, providerHash) {
				pushProvider(provider, providerPath, providerHash, getProcessName(name, 1))
			} else {
				fmt.Println(getProcessName(name, 1), "Already succeed")
			}
		}

	}

}

func pushProvider(key string, path string, hash string, processName string) {
	p := make(map[string]interface{})
	p["key"] = key
	p["path"] = path
	p["hash"] = hash
	jsonP2, _ := json.Marshal(p)
	sAdd(providerQueue, string(jsonP2))
	fmt.Println(processName, "Dispatch", path)
	countToday(providerSet, path)
}
