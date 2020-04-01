package util

import (
	"bytes"
	"encoding/json"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"strconv"
	"time"
)

func status(name string, processNum int) {
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
			"Dists":     getQueueNum(distsKey),
			"Packages":  getQueueNum(packageHashFileKey),
			"Providers": getQueueNum(providerHashFileKey),
		}

		// Statistics
		Content["Statistics"] = map[string]interface{}{
			"Dists_Available":             getSucceedNum(distsKey),
			"Dists_Failed":                getFailedNum(distsKey),
			"Dists_Forbidden":             getStatusCodedFailedNum(distsKey, 403),
			"Dists_Gone":                  getStatusCodedFailedNum(distsKey, 410),
			"Dists_Meta_Missing":          lLen(distsNoMetaKey),
			"Dists_Not_Found":             getStatusCodedFailedNum(distsKey, 404),
			"Dists_Internal_Server_Error": getStatusCodedFailedNum(distsKey, 500),
			"Dists_Bad_Gateway":           getStatusCodedFailedNum(distsKey, 502),
			"Packages":                    lLen(packageKey),
			"Packages_No_Data":            lLen(packagesNoData),
			"Providers":                   lLen(providerKey),
			"Versions":                    lLen(versionsKey),
		}

		// Today Updated
		Content["Today_Updated"] = map[string]interface{}{
			"Dists":             lLen(getTodayKey(distsKey)),
			"Packages":          lLen(getTodayKey(packageKey)),
			"PackagesHashFile":  lLen(getTodayKey(packageHashFileKey)),
			"ProvidersHashFile": lLen(getTodayKey(providerHashFileKey)),
			"Versions":          lLen(getTodayKey(versionsKey)),
		}

		status["Content"] = Content
		status["UpdateAt"] = time.Now().Format("2006-01-02 15:04:05")
		status["CacheSeconds"] = 30

		statusResult, _ := json.Marshal(status)
		options := []oss.Option{
			oss.ContentType("application/json"),
		}

		_ = putObject(getProcessName(name, processNum), "status.json", bytes.NewReader(statusResult), options...)

	}

}
