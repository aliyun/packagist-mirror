package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

var lastTimestamp = ""

func syncV2(processName string) {
	for {
		lastTimestamp = getLastTimestamp()
		if lastTimestamp == "" {
			initTimestamp(processName)
			continue
		}
		getChangesAndUpdateTimestamp(processName)
		// Each cycle requires a time slot
		time.Sleep(time.Duration(config.ApiIterationInterval) * time.Second)
	}
}

func getChangesAndUpdateTimestamp(processName string) {
	url := "metadata/changes.json?since=" + lastTimestamp
	// Get root file from repo
	resp, err := packagistGetApi(url, getProcessName(processName, 1))
	if err != nil {
		return
	}
	// Status code must be 200
	if resp.StatusCode != 200 {
		return
	}
	// Read data stream from body
	content, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Println(getProcessName(processName, 1), packagistUrlApi(url), err.Error())
		return
	}
	// Decode content if Gzip
	content, err = decode(content)
	if err != nil {
		fmt.Println("parseGzip Error", err.Error())
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
	dispatchChanges(changesJson["actions"], processName)
	setLastTimestamp(timestampAPI)
}

func initTimestamp(processName string) {
	// Get root file from repo
	resp, err := packagistGetApi("metadata/changes.json", getProcessName(processName, 1))
	if err != nil {
		return
	}
	// Status code must be 400
	if resp.StatusCode != 400 {
		return
	}
	// Read data stream from body
	content, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Println(getProcessName(processName, 1), packagistUrlApi("metadata/changes.json"), err.Error())
		return
	}
	// Decode content if Gzip
	content, err = decode(content)
	if err != nil {
		fmt.Println("parseGzip Error", err.Error())
		return
	}
	// JSON Decode
	changesJson := make(map[string]interface{})
	err = json.Unmarshal(content, &changesJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	syncAll(processName)
	timestampAPI := strconv.FormatInt(int64(changesJson["timestamp"].(float64)), 10)
	setLastTimestamp(timestampAPI)
}

func syncAll(processName string) {
	// Get root file from repo
	resp, err := packagistGetApi("packages/list.json", getProcessName(processName, 1))
	if err != nil {
		return
	}
	// Status code must be 200
	if resp.StatusCode != 200 {
		return
	}
	// Read data stream from body
	content, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		fmt.Println(getProcessName(processName, 1), packagistUrlApi("packages/list.json"), err.Error())
		return
	}
	// Decode content if Gzip
	content, err = decode(content)
	if err != nil {
		fmt.Println("parseGzip Error", err.Error())
		return
	}
	// JSON Decode
	list := make(map[string][]string)
	err = json.Unmarshal(content, &list)
	if err != nil {
		fmt.Println(getProcessName(processName, 1), err.Error())
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

func dispatchChanges(changes interface{}, name string) {
	for _, item := range changes.([]interface{}) {

		actionType := item.(map[string]interface{})["type"].(string)
		packageName := item.(map[string]interface{})["package"].(string)
		updateTime := strconv.FormatInt(int64(item.(map[string]interface{})["time"].(float64)), 10)
		if !hGetValue(packageV2Set, packageName, updateTime) {
			pushV2Queue(actionType, packageName, updateTime)
		} else {
			fmt.Println(getProcessName(name, 1), "File is up to date:", packageName, updateTime)
		}

	}

}
