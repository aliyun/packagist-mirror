package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func (ctx *Context) SyncV2(processName string) {
	var logger = log.New(os.Stderr, processName, log.LUTC)
	for {
		err := syncV2(ctx)
		if err != nil {
			logger.Println("sync v2 failed: " + err.Error())
			continue
		}

		// Each cycle requires a time slot
		time.Sleep(time.Duration(ctx.mirror.apiIterationInterval) * time.Second)
	}
}

func syncV2(ctx *Context) (err error) {
	lastTimestamp, err := ctx.redis.Get("lastTimestamp").Result()
	if err != nil {
		err = fmt.Errorf("get last timestamp failed" + err.Error())
		return
	}

	if lastTimestamp == "" {
		lastTimestamp, err = ctx.packagist.GetInitTimestamp()
		if err != nil {
			return
		}

		syncAll(ctx)

		err = ctx.redis.Set("lastTimestamp", lastTimestamp, 0).Err()
		return
	}

	changes, err := ctx.packagist.GetMetadataChanges(lastTimestamp)
	if err != nil {
		return
	}

	// Dispatch changes
	timestampAPI := strconv.FormatInt(int64(changes.Timestamp), 10)
	if timestampAPI == lastTimestamp {
		// No changes
		return
	}

	for _, item := range changes.Actions {
		packageName := item.Package
		updateTime := strconv.FormatInt(int64(item.Time), 10)
		storedUpdateTime, err2 := ctx.redis.HGet(packageV2Set, packageName).Result()
		if err2 != nil {
			err = fmt.Errorf("get time failed: " + err2.Error())
			return
		}

		if storedUpdateTime != updateTime {
			ctx.redis.SAdd(packageV2Queue, item.ToJSONString()).Result()
		} else {
			fmt.Println("File is up to date:", packageName, updateTime)
		}
	}

	err = ctx.redis.Set("lastTimestamp", timestampAPI, 0).Err()
	return
}

func syncAll(ctx *Context) (err error) {
	// Get root file from repo
	content, err := ctx.packagist.GetAllPackages()
	if err != nil {
		return
	}

	// JSON Decode
	list := make(map[string][]string)
	err = json.Unmarshal(content, &list)
	if err != nil {
		return
	}

	for _, packageName := range list["packageNames"] {
		_, err = ctx.redis.SAdd(packageV2Queue, NewChangeAction("update", packageName, 0).ToJSONString()).Result()
		if err != nil {
			return
		}

		_, err = ctx.redis.SAdd(packageV2Queue, NewChangeAction("update", packageName+"~dev", 0).ToJSONString()).Result()
		if err != nil {
			return
		}
	}

	return
}
