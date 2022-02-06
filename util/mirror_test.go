package util

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMirror(t *testing.T) {
	mirror := NewMirror("providerUrl", "distUrl", 5)
	assert.Equal(t, "providerUrl", mirror.providerUrl)
}

func TestChanges(t *testing.T) {
	var jsonStr = "{\"actions\":[{\"type\":\"update\",\"package\":\"codeception\\/module-datafactory~dev\",\"time\":1644160111}],\"timestamp\":16441601329960}"
	changes := new(Changes)
	json.Unmarshal([]byte(jsonStr), &changes)
	assert.Equal(t, 16441601329960, changes.Timestamp)
	assert.Equal(t, 1, len(changes.Actions))
	action := changes.Actions[0]
	assert.Equal(t, "codeception/module-datafactory~dev", action.Package)
	assert.Equal(t, 1644160111, action.Time)
	assert.Equal(t, "update", action.Type)
}
