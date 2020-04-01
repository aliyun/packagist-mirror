package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func dists(name string, num int) {
	for {
		job, err := popFromQueue(distsKey)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		uploadDist(job[1], name, num)
	}

}

func distsRetry(statusCode int, num int) {

	name := "distsRetry" + strconv.Itoa(statusCode)

	for {
		job, err := popFromQueueStatusCode(distsKey, statusCode)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Println(getProcessName(name, num), "Dist Retry", statusCode, job)

		uploadDist(job[1], name, num)
	}

}

func uploadDist(jobJson string, name string, num int) {

	removeFromQueue(distsKey, jobJson, getProcessName(name, num))

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
	count(distsKey, path)

	// Redis IsObjectExist
	if isSucceed(distsKey, path) {
		fmt.Println(getProcessName(name, num), "Succeed", mirrorUrl(path))
		return
	}

	// OSS IsObjectExist
	if isObjectExist(getProcessName(name, num), path) {
		makeSucceed(distsKey, path, getProcessName(name, num))
		return
	}

	// Get dist
	fmt.Println(getProcessName(name, num), "Downloading", url)
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
		makeFailed(distsKey, path, jobJson+err.Error())
		return
	}

	if resp.StatusCode != 200 {
		syncHasError = true
		// Make failed count
		makeStatusCodeFailed(distsKey, resp.StatusCode, path, url)

		// Push into failed queue to retry
		if resp.StatusCode != 404 && resp.StatusCode != 410 {
			pushToQueueForStatusCodeRetry(distsKey, resp.StatusCode, jobJson)
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

	makeSucceed(distsKey, path, getProcessName(name, num))

	// Build Cache for DNS
	fmt.Println(getProcessName(name, num), "Build Cache for DNS")
	resp, _ = mirrorGet(path, getProcessName(name, num))
	_ = resp.Body.Close()

}
