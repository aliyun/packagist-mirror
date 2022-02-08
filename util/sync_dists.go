package util

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func (ctx *Context) SyncDists(processName string) {
	var logger = NewLogger(processName)
	for {
		jobJson, err := ctx.redis.SPop(distQueue).Result()

		if err == redis.Nil {
			// logger.Info("get no task from " + distQueue + ", sleep 1 second")
			time.Sleep(1 * time.Second)
			continue
		}

		dist, err := NewDistFromJSONString(jobJson)
		if err != nil {
			logger.Error("Covert to Dist failed: " + jobJson)
			time.Sleep(1 * time.Second)
			continue
		}

		logger.Info("upload dist for " + dist.Path)
		err = uploadDist(ctx, logger, dist)
		if err != nil {
			logger.Error(fmt.Sprintf("sync dist(%s) failed. ", jobJson) + err.Error())
			continue
		}
	}

}

func (ctx *Context) SyncDistsRetry(processName string) {
	var logger = NewLogger(processName)
	for {
		jobJson, err := ctx.redis.SPop(distQueueRetry).Result()
		if jobJson == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		dist, err := NewDistFromJSONString(jobJson)
		if err != nil {
			logger.Error("Covert to Dist failed: " + jobJson)
			time.Sleep(1 * time.Second)
			continue
		}

		err = uploadDist(ctx, logger, dist)
		if err != nil {
			logger.Error("sync dist failed. " + err.Error())
			continue
		}
	}

}

func uploadDist(ctx *Context, logger *MyLogger, job *Dist) (err error) {
	// Get information
	path := job.Path
	url := job.Url

	if url == "" {
		err = fmt.Errorf("url is invalid")
		return
	}

	// Count
	ctx.redis.SAdd(getTodayKey(distSet), path).Result()
	ctx.redis.ExpireAt(getTodayKey(distSet), getTomorrow())

	// OSS IsObjectExist
	isExist, err := ctx.ossBucket.IsObjectExist(path)
	if err != nil {
		return
	}

	if isExist {
		return
	}

	// Get dist
	resp, err := ctx.github.GetDist(url)

	if err != nil {
		// makeFailed(distSet, path, err)
		return
	}

	// if resp.StatusCode != 200 {
	// 	syncHasError = true
	// 	// Make failed count
	// 	makeStatusCodeFailed(distSet, resp.StatusCode, url)

	// 	// Push into failed queue to retry
	// 	if resp.StatusCode != 404 && resp.StatusCode != 410 {
	// 		sAdd(distQueueRetry, job.ToJSONString())
	// 	}

	// 	return
	// }

	// Put into OSS
	err = ctx.ossBucket.PutObject(path, resp.Body)
	if err != nil {
		return
	}

	// makeSucceed(distSet, path)
	ctx.cdn.WarmUp(path)
	return
}
