package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func (ctx *Context) SyncV2(processName string) {
	var logger = NewLogger(processName)
	for {
		err := syncV2(ctx, logger)
		if err != nil {
			logger.Error("sync v2 failed: " + err.Error())
			continue
		}

		// Each cycle requires a time slot
		time.Sleep(time.Duration(ctx.mirror.apiIterationInterval) * time.Second)
	}
}

func syncV2(ctx *Context, logger *MyLogger) (err error) {

	lastTimestamp, err := ctx.redis.Get(v2LastUpdateTimeKey).Result()
	if err != nil && err != redis.Nil {
		err = fmt.Errorf("get last timestamp failed" + err.Error())
		return
	}

	if lastTimestamp == "" {
		lastTimestamp, err = ctx.packagist.GetInitTimestamp()
		if err != nil {
			return
		}

		err = syncAll(ctx, logger)
		if err != nil {
			return
		}
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
		if err2 == redis.Nil {
			ctx.redis.SAdd(packageV2Queue, item.ToJSONString()).Result()
			continue
		}

		if err2 != nil {
			err = fmt.Errorf("get time failed: " + err2.Error())
			return
		}

		if storedUpdateTime != updateTime {
			ctx.redis.SAdd(packageV2Queue, item.ToJSONString()).Result()
		}
	}

	err = ctx.redis.Set(v2LastUpdateTimeKey, timestampAPI, 0).Err()
	return
}

func syncAll(ctx *Context, logger *MyLogger) (err error) {
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
