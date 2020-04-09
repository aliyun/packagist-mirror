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

func packagesJsonFile(name string, num int) {

	processName := getProcessName(name, num)

	for {
		// Each cycle requires a time slot
		time.Sleep(1 * time.Second)

		// Get root file from repo
		resp, err := packagistGet("packages.json", processName)
		if err != nil {
			continue
		}

		// Status code must be 200
		if resp.StatusCode != 200 {
			continue
		}

		// Get Last-Modified field
		packagistLastModified = resp.Header["Last-Modified"][0]

		// Read data stream from body
		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Println(processName, packagistUrl("packages.json"), err.Error())
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
			fmt.Println(processName, "Update to date: packages.json")
			continue
		}
		packagesJsonCache = content

		// JSON Decode
		err = json.Unmarshal(content, &packagesJson)
		if err != nil {
			errHandler(err)
			continue
		}

		// Make error false
		syncHasError = false

		// Dispatch providers
		dispatchProviders(packagesJson["provider-includes"], processName)

		for {
			// If all tasks are completed, skip the loop and update the file
			if queueExists(providerHashFileKey) == 0 && queueExists(packageHashFileKey) == 0 && queueExists(distsKey) == 0 {
				break
			}
			fmt.Println(processName, "Synchronization task is not completed, check again in 1 second.")
			time.Sleep(1 * time.Second)
		}

		if syncHasError == true {
			fmt.Println(processName, "There is an error in this synchronization. We look forward to the next synchronization...")
			continue
		}

		time.Sleep(5 * time.Second)

		// Update `packages.json`
		packagesJson["last-update"] = time.Now().Format("2006-01-02 15:04:05")
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
		_ = putObject(processName, "packages.json", contentReader, options...)
	}

}

func dispatchProviders(distMap interface{}, processName string) {

	for provider, value := range distMap.(map[string]interface{}) {

		count(providerKey, provider)

		for _, hash := range value.(map[string]interface{}) {

			path := strings.Replace(provider, "%hash%", hash.(string), -1)

			if isSucceed(providerHashFileKey, path) {
				fmt.Println(processName, "Already succeed", mirrorUrl(path))
				continue
			}

			pushProviderToQueue(path, processName)
		}

	}

}

func pushProviderToQueue(path string, processName string) {
	if pushToQueue(providerHashFileKey, path, processName) {
		count(providerHashFileKey, path)
		fmt.Println(processName, "Dispatch", path)
	}
}
