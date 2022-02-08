package util

import "encoding/json"

type Mirror struct {
	providerUrl          string
	distUrl              string
	apiIterationInterval int
}

func NewMirror(providerUrl string, distUrl string, apiIterationInterval int) (mirror *Mirror) {
	return &Mirror{
		providerUrl:          providerUrl,
		distUrl:              distUrl,
		apiIterationInterval: apiIterationInterval,
	}
}

type Dist struct {
	Path string `json:"path"`
	Url  string `json:"url"`
}

func NewDist(path, url string) *Dist {
	return &Dist{Path: path, Url: url}
}

func NewDistFromJSONString(jsonString string) (dist *Dist, err error) {
	dist = new(Dist)
	err = json.Unmarshal([]byte(jsonString), dist)
	return
}

func (dist *Dist) ToJSONString() string {
	distString, _ := json.Marshal(dist)
	return string(distString)
}

type Changes struct {
	Timestamp int            `json:"timestamp"`
	Actions   []ChangeAction `json:"actions"`
}

type ChangeAction struct {
	Type    string `json:"type"`
	Package string `json:"package"`
	Time    int    `json:"time"`
}

func NewChangeAction(type_ string, packageName string, time int) *ChangeAction {
	return &ChangeAction{
		Type:    type_,
		Package: packageName,
		Time:    time,
	}
}

func NewChangeActionFromJSONString(jsonString string) (action *ChangeAction, err error) {
	action = new(ChangeAction)
	err = json.Unmarshal([]byte(jsonString), action)
	return
}

func (action *ChangeAction) ToJSONString() string {
	jsonStr, _ := json.Marshal(action)
	return string(jsonStr)
}

type Task struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func NewTask(key, path, hash string) *Task {
	return &Task{
		Key:  key,
		Path: path,
		Hash: hash,
	}
}

func NewTaskFromJSONString(jsonString string) (task *Task, err error) {
	task = new(Task)
	err = json.Unmarshal([]byte(jsonString), task)
	return
}

type Providers struct {
	Providers map[string]Hashes `json:"providers"`
}

func NewProvidersFromJSONString(jsonString string) (providers *Providers, err error) {
	providers = new(Providers)
	err = json.Unmarshal([]byte(jsonString), providers)
	return
}
