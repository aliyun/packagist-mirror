package util

import (
	"encoding/json"
	"fmt"
	"time"
)

func (ctx *Context) SyncDists(processName string) {

	for {
		jobJson := sPop(distQueue)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		ctx.uploadDist(jobJson)
	}

}

func (ctx *Context) SyncDistsRetry(processName string) {

	for {
		jobJson := sPop(distQueueRetry)
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(2 * time.Second)
		ctx.uploadDist(jobJson)
	}

}

func (ctx *Context) uploadDist(jobJson string) {
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
		fmt.Println("Dist field not found: path")
		return
	}

	url, ok := distMap["url"]
	if !ok {
		fmt.Println("Dist field not found: url")
		return
	}

	// Count
	countToday(distSet, path)

	// OSS IsObjectExist
	if isExist, _ := ctx.ossBucket.IsObjectExist(path); isExist {
		makeSucceed(distSet, path)
		return
	}

	// Get dist
	resp, err := ctx.github.GetDist(url)

	if err != nil {
		syncHasError = true
		fmt.Println(path, err.Error())
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
			"Dist Get Error",
			resp.StatusCode,
			jobJson,
		)
		return
	}
	// Put into OSS
	err = ctx.ossBucket.PutObject(path, resp.Body)
	if err != nil {
		syncHasError = true
		return
	}

	makeSucceed(distSet, path)
	ctx.cdn.WarmUp(path)
}
