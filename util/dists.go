package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func dists(name string, num int) {

	for {
		jobJson := sPop(distQueue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		uploadDist(jobJson, name, num)
	}

}

func distsRetry(name string, num int) {

	for {
		jobJson := sPop(distQueueRetry)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		uploadDist(jobJson, name, num)
	}

}

func uploadDist(jobJson string, name string, num int) {

	// Json decode
	distMap := make(map[string]string)
	err := json.Unmarshal([]byte(jobJson), &distMap)
	if err != nil {
		errHandler(err)
		return
	}

	// Get information
	path, ok := distMap["path"]
	if !ok {
		fmt.Println(getProcessName(name, num), "Dist field not found: path")
		return
	}

	url, ok := distMap["url"]
	if !ok {
		fmt.Println(getProcessName(name, num), "Dist field not found: url")
		return
	}

	// Count
	countToday(distSet, path)

	// OSS IsObjectExist
	if isObjectExist(getProcessName(name, num), path) {
		makeSucceed(distSet, path)
		return
	}

	// Get dist
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	// add authorization header to the req
	req.Header.Add("Authorization", "token "+config.GithubToken)
	req.Header.Add("User-Agent", "Alibaba")
	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		syncHasError = true
		fmt.Println(getProcessName(name, num), path, err.Error())
		makeFailed(distSet, path, err)
		return
	}

	if resp.StatusCode != 200 {
		syncHasError = true
		// Make failed count
		makeStatusCodeFailed(distSet, resp.StatusCode, url)

		// Push into failed queue to retry
		if resp.StatusCode != 404 && resp.StatusCode != 410 {
			sAdd(distQueueRetry, jobJson)
		}

		fmt.Println(
			getProcessName(name, num),
			"Dist Get Error",
			resp.StatusCode,
			jobJson,
		)
		return
	}

	// Put into OSS
	err = putObject(getProcessName(name, num), path, resp.Body)
	if err != nil {
		syncHasError = true
		return
	}

	makeSucceed(distSet, path)

	cdnCache(path, name, num)
}
