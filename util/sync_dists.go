package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

func (ctx *Context) SyncDists(processName string) {
	var logger = log.New(os.Stderr, processName, log.LUTC)
	for {
		jobJson, err := ctx.redis.SPop(distQueue).Result()
		if err != nil {
			logger.Println("pop from queue(" + distQueue + ") failed")
			time.Sleep(1 * time.Second)
			continue
		}

		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		dist, err := NewDistFromJSONString(jobJson)
		if err != nil {
			logger.Println("Covert to Dist failed: " + jobJson)
			time.Sleep(1 * time.Second)
			continue
		}

		uploadDist(ctx, dist)
	}

}

func (ctx *Context) SyncDistsRetry(processName string) {
	var logger = log.New(os.Stderr, processName, log.LUTC)
	for {
		jobJson, err := ctx.redis.SPop(distQueueRetry).Result()
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(2 * time.Second)

		dist, err := NewDistFromJSONString(jobJson)
		if err != nil {
			logger.Println("Covert to Dist failed: " + jobJson)
			time.Sleep(1 * time.Second)
			continue
		}

		uploadDist(ctx, dist)
	}

}

func uploadDist(ctx *Context, job *Dist) (err error) {
	// Get information
	path := job.Path
	url := job.Url

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
			sAdd(distQueueRetry, job.ToJSONString())
		}

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
	return
}
