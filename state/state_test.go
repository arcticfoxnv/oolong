package state

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestNewFileState(t *testing.T) {
	state := NewFileState("test.json").(*fileState)
	if state.Filename != "test.json" {
		t.Fail()
	}

	if state.LastUpdated == nil {
		t.Fail()
	}
}

func TestNewStateFromFile(t *testing.T) {
	testData := `{"AccessToken": "abc"}`
	ioutil.WriteFile("test.json", []byte(testData), 0600)

	stateIface, err := NewStateFromFile("test.json")
	if err != nil {
		t.FailNow()
	}
	state := stateIface.(*fileState)

	if state.Filename != "test.json" {
		t.Fail()
	}

	if state.AccessToken != "abc" {
		t.Fail()
	}

	os.Remove("test.json")
}

func TestNewStateFromFileMissingFile(t *testing.T) {
	state, err := NewStateFromFile("test.json")
	if err == nil {
		t.Fail()
	}

	if state != nil {
		t.Fail()
	}
}

func TestNewStateFromFileCorruptFile(t *testing.T) {
	testData := `"garbage`
	ioutil.WriteFile("test.json", []byte(testData), 0600)

	state, err := NewStateFromFile("test.json")
	if err == nil {
		t.Fail()
	}

	if state != nil {
		t.Fail()
	}

	os.Remove("test.json")
}

func TestFileStateFirstUpdate(t *testing.T) {
	state := NewFileState("test.json")
	now := time.Now()
	state.Update("xxx", "test", now)
	updateTime := state.GetLastUpdateTime("xxx", "test")
	if !updateTime.Equal(now) {
		t.Fail()
	}
}

func TestFileStateSecondUpdate(t *testing.T) {
	state := NewFileState("test.json")
	now := time.Now()
	state.Update("xxx", "test", now)

	prev := now.Add(-5 * time.Minute)
	state.Update("xxx", "test", prev)

	updateTime := state.GetLastUpdateTime("xxx", "test")
	if updateTime.Equal(now) {
		t.Fail()
	}
}

func TestFileStateGetLastUptimeTimeNil(t *testing.T) {
	state := NewFileState("test.json")
	lastUpdate := state.GetLastUpdateTime("xxx", "test")
	if !lastUpdate.Equal(time.Time{}) {
		t.Fail()
	}
}

func TestFileStateAccessToken(t *testing.T) {
	state := NewFileState("test.json")
	state.SetAccessToken("xxx")
	if state.GetAccessToken() != "xxx" {
		t.Fail()
	}
}

func TestFileStateSave(t *testing.T) {
	state := NewFileState("test.json")
	state.SetAccessToken("xxx")
	state.Save()

	stateIface, err := NewStateFromFile("test.json")
	if err != nil {
		t.FailNow()
	}
	stateFromFile := stateIface.(*fileState)
	if stateFromFile.GetAccessToken() != "xxx" {
		t.Fail()
	}

	os.Remove("test.json")
}
