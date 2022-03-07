package util

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
)

func (ctx *Context) SyncStatus(processName string) {
	var logger = NewLogger(processName)
	logger.Info("start to sync status")
	for {
		time.Sleep(1 * time.Second)

		// Initialize variables
		status := make(map[string]interface{})
		content := make(map[string]interface{})

		// If this variable is equal to an empty string
		// it may be that the other coroutine has not yet obtained
		// proceeds to the next loop
		packagistLastModified, err := ctx.redis.Get(packagistLastModifiedKey).Result()
		if err == redis.Nil {
			continue
		}

		if err != nil {
			logger.Error(err.Error())
			continue
		}

		// Format: 2006-01-02 15:04:05
		aliDateTime, err := ctx.redis.Get(lastUpdateTimeKey).Result()
		if err == redis.Nil {
			continue
		}

		if err != nil {
			logger.Error(err.Error())
			continue
		}

		loc, _ := time.LoadLocation("Asia/Shanghai")
		packagistLast, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", packagistLastModified)
		packagistDateTime := packagistLast.In(loc).Format("2006-01-02 15:04:05")

		aliComposerLast, _ := time.Parse("2006-01-02 15:04:05", aliDateTime)
		packagistLast, _ = time.Parse("2006-01-02 15:04:05", packagistDateTime)

		interval := int(aliComposerLast.In(loc).Unix() - packagistLast.In(loc).Unix())
		status["Delayed"] = 0
		status["Interval"] = interval
		if interval < 0 {
			status["Title"] = "Delayed " + strconv.Itoa(-interval) + " Seconds, waiting for updates..."
			status["Delayed"] = -interval
			if -interval >= 600 {
				status["ShouldReportDelay"] = true
			}
		} else {
			status["Title"] = "Synchronized within " + strconv.Itoa(interval) + " Seconds!"
			status["ShouldReportDelay"] = false
		}

		content["Last_Update"] = map[string]interface{}{
			"AliComposer": aliDateTime,
			"Packagist":   packagistDateTime,
		}

		// Queue
		content["Queue"] = map[string]interface{}{
			"Providers":  ctx.redisSCard(providerQueue),
			"Packages":   ctx.redisSCard(packageP1Queue) + ctx.redisSCard(packageV2Queue),
			"Dists":      ctx.redisSCard(distQueue),
			"DistsRetry": ctx.redisSCard(distQueueRetry),
		}

		// Statistics
		content["Statistics"] = map[string]interface{}{
			"Dists_Available":             ctx.redisSCard(distSet),
			"Dists_Failed":                ctx.countFailed(distSet),
			"Dists_Forbidden":             ctx.countStatusCodedFailed(distSet, 403),
			"Dists_Gone":                  ctx.countStatusCodedFailed(distSet, 410),
			"Dists_Meta_Missing":          ctx.redisSCard(distsNoMetaKey),
			"Dists_Not_Found":             ctx.countStatusCodedFailed(distSet, 404),
			"Dists_Internal_Server_Error": ctx.countStatusCodedFailed(distSet, 500),
			"Dists_Bad_Gateway":           ctx.countStatusCodedFailed(distSet, 502),
			"Packages":                    ctx.redisHLen(packageV1Set),
			"Packages_No_Data":            ctx.redisSCard(packagesNoData),
			"Providers":                   ctx.redisHLen(providerSet),
			"Versions":                    ctx.redisSCard(versionsSet),
		}

		// Today Updated
		content["Today_Updated"] = map[string]interface{}{
			"Dists":             ctx.redisSCard(getTodayKey(distSet)),
			"Packages":          ctx.redisSCard(getTodayKey(packageV1Set)),
			"PackagesHashFile":  ctx.redisSCard(getTodayKey(packageV1SetHash)),
			"ProvidersHashFile": ctx.redisSCard(getTodayKey(providerSet)),
			"Versions":          ctx.redisSCard(getTodayKey(versionsSet)),
		}

		status["Content"] = content
		status["CacheSeconds"] = 30

		// Update status.json
		status["UpdateAt"] = time.Now().Format("2006-01-02 15:04:05")
		ossStatusContent, _ := json.Marshal(status)
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		_ = ctx.ossBucket.PutObject("status.json", bytes.NewReader(ossStatusContent), options...)
	}

}
