package util

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
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
		if packagistLastModified == "" {
			continue
		}

		// Format: 2006-01-02 15:04:05
		aliDateTime, err := ctx.redis.Get(lastUpdateTimeKey).Result()
		if err != nil {
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
			"Providers":  sCard(providerQueue),
			"Packages":   sCard(packageP1Queue) + sCard(packageV2Queue),
			"Dists":      sCard(distQueue),
			"DistsRetry": sCard(distQueueRetry),
		}

		// Statistics
		content["Statistics"] = map[string]interface{}{
			"Dists_Available":             sCard(distSet),
			"Dists_Failed":                countFailed(distSet),
			"Dists_Forbidden":             countStatusCodedFailed(distSet, 403),
			"Dists_Gone":                  countStatusCodedFailed(distSet, 410),
			"Dists_Meta_Missing":          sCard(distsNoMetaKey),
			"Dists_Not_Found":             countStatusCodedFailed(distSet, 404),
			"Dists_Internal_Server_Error": countStatusCodedFailed(distSet, 500),
			"Dists_Bad_Gateway":           countStatusCodedFailed(distSet, 502),
			"Packages":                    hLen(packageV1Set),
			"Packages_No_Data":            sCard(packagesNoData),
			"Providers":                   hLen(providerSet),
			"Versions":                    sCard(versionsSet),
		}

		// Today Updated
		content["Today_Updated"] = map[string]interface{}{
			"Dists":             sCard(getTodayKey(distSet)),
			"Packages":          sCard(getTodayKey(packageV1Set)),
			"PackagesHashFile":  sCard(getTodayKey(packageV1SetHash)),
			"ProvidersHashFile": sCard(getTodayKey(providerSet)),
			"Versions":          sCard(getTodayKey(versionsSet)),
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
