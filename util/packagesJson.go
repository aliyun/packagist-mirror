package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"strings"
	"time"
)

var packagesJson = make(map[string]interface{})
var packagesJsonCache []byte
var packagistLastModified = ""
var syncHasError = false

func packagesJsonFile(name string) {

	for {
		// Each cycle requires a time slot
		time.Sleep(1 * time.Second)

		// Get root file from repo
		resp, err := packagistGet("packages.json", getProcessName(name, 1))
		if err != nil {
			sAdd(packagesJsonKey+"-Get", "packages.json")
			continue
		}

		// Status code must be 200
		if resp.StatusCode != 200 {
			makeStatusCodeFailed(packagesJsonKey, resp.StatusCode, packagistUrl("packages.json"))
			continue
		}

		// Get Last-Modified field
		packagistLastModified = resp.Header["Last-Modified"][0]

		// Read data stream from body
		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Println(getProcessName(name, 1), packagistUrl("packages.json"), err.Error())
			continue
		}

		// Decode content if Gzip
		content, err = decode(content)
		if err != nil {
			fmt.Println("parseGzip Error", err.Error())
			continue
		}

		// Cache content
		if bytes.Equal(packagesJsonCache, content) {
			fmt.Println(getProcessName(name, 1), "Update to date: packages.json")
			continue
		}
		packagesJsonCache = content

		// JSON Decode
		err = json.Unmarshal(content, &packagesJson)
		if err != nil {
			sAdd("root-json_decode_error", "root")
			continue
		}

		// Make error false
		syncHasError = false

		// Dispatch providers
		dispatchProviders(packagesJson["provider-includes"], name)

		for {
			// If all tasks are completed, skip the loop and update the file
			left := sCard(distQueue) + sCard(providerQueue) + sCard(packageP1Queue) + sCard(packageP2Queue)
			if left == 0 {
				break
			}
			fmt.Println(getProcessName(name, 1), "Processing:", left, ", Check again in 1 second. ")
			time.Sleep(1 * time.Second)
		}

		if syncHasError == true {
			fmt.Println(getProcessName(name, 1), "There is an error in this synchronization. We look forward to the next synchronization...")
			continue
		}

		// Update `packages.json`
		packagesJson["last-update"] = time.Now().Format("2006-01-02 15:04:05")
		packagesJson["metadata-url"] = config.ProviderUrl + "p2/%package%.json"
		packagesJson["providers-url"] = config.ProviderUrl + "p/%package%$%hash%.json"
		packagesJson["mirrors"] = []map[string]interface{}{
			{
				"dist-url":  config.DistUrl + "dists/%package%/%reference%.%type%",
				"preferred": true,
			},
		}

		// Json Encode
		content, _ = json.Marshal(packagesJson)
		contentReader := bytes.NewReader(content)
		options := []oss.Option{
			oss.ContentType("application/json"),
		}

		// Upload Content
		_ = putObject(getProcessName(name, 1), "packages.json", contentReader, options...)
	}

}

func dispatchProviders(distMap interface{}, name string) {

	for provider, value := range distMap.(map[string]interface{}) {

		for _, hash := range value.(map[string]interface{}) {

			providerHash := hash.(string)
			providerPath := strings.Replace(provider, "%hash%", providerHash, -1)

			if hashHGet(providerSet, provider, providerHash) {
				fmt.Println(getProcessName(name, 1), "Already succeed", mirrorUrl(providerPath))
			} else {
				pushProvider(provider, providerPath, providerHash, getProcessName(name, 1))
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
