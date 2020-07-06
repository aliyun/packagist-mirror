package util

import (
	"bytes"
	"encoding/json"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"strconv"
	"time"
)

var statusContentCache []byte

func status(name string, processNum int) {

	processName := getProcessName(name, processNum)

	for {
		time.Sleep(1 * time.Second)

		// Initialize variables
		status := make(map[string]interface{})
		Content := make(map[string]interface{})

		// If this variable is equal to an empty string
		// it may be that the other coroutine has not yet obtained
		// proceeds to the next loop
		if packagistLastModified == "" {
			continue
		}

		// Format: 2006-01-02 15:04:05
		aliDateTime, ok := packagesJson["last-update"].(string)
		if !ok {
			continue
		}

		Loc, _ := time.LoadLocation("Asia/Shanghai")
		packagistLast, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", packagistLastModified)
		packagistDateTime := packagistLast.In(Loc).Format("2006-01-02 15:04:05")

		aliComposerLast, _ := time.Parse("2006-01-02 15:04:05", aliDateTime)
		packagistLast, _ = time.Parse("2006-01-02 15:04:05", packagistDateTime)

		Interval := int(aliComposerLast.In(Loc).Unix() - packagistLast.In(Loc).Unix())
		status["Delayed"] = 0
		status["Interval"] = Interval
		if Interval < 0 {
			status["Title"] = "Delayed " + strconv.Itoa(-Interval) + " Seconds, waiting for updates..."
			status["Delayed"] = -Interval
			if -Interval >= 600 {
				status["ShouldReportDelay"] = true
			}
		} else {
			status["Title"] = "Synchronized within " + strconv.Itoa(Interval) + " Seconds!"
			status["ShouldReportDelay"] = false
		}

		Content["Last_Update"] = map[string]interface{}{
			"AliComposer": aliDateTime,
			"Packagist":   packagistDateTime,
		}

		// Queue
		Content["Queue"] = map[string]interface{}{
			"Providers":  sCard(providerQueue),
			"Packages":   sCard(packageP1Queue) + sCard(packageV2Queue),
			"Dists":      sCard(distQueue),
			"DistsRetry": sCard(distQueueRetry),
		}

		// Statistics
		Content["Statistics"] = map[string]interface{}{
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
		Content["Today_Updated"] = map[string]interface{}{
			"Dists":             sCard(getTodayKey(distSet)),
			"Packages":          sCard(getTodayKey(packageV1Set)),
			"PackagesHashFile":  sCard(getTodayKey(packageV1SetHash)),
			"ProvidersHashFile": sCard(getTodayKey(providerSet)),
			"Versions":          sCard(getTodayKey(versionsSet)),
		}

		status["Content"] = Content
		status["CacheSeconds"] = 30

		statusContent, _ := json.Marshal(status)

		if bytes.Equal(statusContentCache, statusContent) {
			continue
		}

		// Update status.json
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		err := putObject(processName, "status.json", bytes.NewReader(statusContent), options...)
		if err != nil {
			continue
		}

		// The cache is updated only if the push is successful
		statusContentCache = statusContent
	}

}
