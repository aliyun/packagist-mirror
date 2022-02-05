package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

var lastTimestamp = ""

func (ctx *Context) SyncV2(processName string) {
	for {
		lastTimestamp = getLastTimestamp()
		if lastTimestamp == "" {
			ctx.initTimestamp()
			continue
		}
		ctx.getChangesAndUpdateTimestamp()
		// Each cycle requires a time slot
		time.Sleep(time.Duration(ctx.mirror.apiIterationInterval) * time.Second)
	}
}

func (ctx *Context) getChangesAndUpdateTimestamp() {
	// Get root file from repo
	content, err := ctx.packagist.GetMetadataChanges(lastTimestamp)
	if err != nil {
		// TODO
		return
	}

	// JSON Decode
	changesJson := make(map[string]interface{})
	err = json.Unmarshal(content, &changesJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Dispatch changes
	timestampAPI := strconv.FormatInt(int64(changesJson["timestamp"].(float64)), 10)
	if timestampAPI == lastTimestamp {
		return
	}
	dispatchChanges(changesJson["actions"])
	setLastTimestamp(timestampAPI)
}

func (ctx *Context) initTimestamp() {
	// Get root file from repo
	content, err := ctx.packagist.GetInitMetadataChanges()
	if err != nil {
		// TODO:
		return
	}

	// JSON Decode
	changesJson := make(map[string]interface{})
	err = json.Unmarshal(content, &changesJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ctx.syncAll()

	timestampAPI := strconv.FormatInt(int64(changesJson["timestamp"].(float64)), 10)
	setLastTimestamp(timestampAPI)
}

func (ctx *Context) syncAll() {
	// Get root file from repo
	content, err := ctx.packagist.GetAllPackages()

	if err != nil {
		// TODO
		return
	}

	// JSON Decode
	list := make(map[string][]string)
	err = json.Unmarshal(content, &list)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, packageName := range list["packageNames"] {
		pushV2Queue("update", packageName, "0")
		pushV2Queue("update", packageName+"~dev", "0")
	}

}

func pushV2Queue(actionType string, packageName string, time string) {
	v2 := make(map[string]interface{})
	v2["type"] = actionType
	v2["package"] = packageName
	v2["time"] = time
	jsonV2, _ := json.Marshal(v2)
	sAdd(packageV2Queue, string(jsonV2))
}

func dispatchChanges(changes interface{}) {
	for _, item := range changes.([]interface{}) {

		actionType := item.(map[string]interface{})["type"].(string)
		packageName := item.(map[string]interface{})["package"].(string)
		updateTime := strconv.FormatInt(int64(item.(map[string]interface{})["time"].(float64)), 10)
		if !hGetValue(packageV2Set, packageName, updateTime) {
			pushV2Queue(actionType, packageName, updateTime)
		} else {
			fmt.Println("File is up to date:", packageName, updateTime)
		}

	}

}
