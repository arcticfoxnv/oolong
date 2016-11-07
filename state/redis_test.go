package state

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gopkg.in/redis.v5"
)

const testRedisHost = "localhost"
const testRedisPort = 6379
const testRedisKey = "oolong.test"

func TestNewRedisState(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey).(*redisState)
	if state.client == nil {
		t.Fail()
	}

	if state.key != testRedisKey {
		t.Fail()
	}

	if state.LastUpdated == nil {
		t.Fail()
	}
}

func TestNewStateFromRedis(t *testing.T) {
	testData := `{"AccessToken": "abc"}`
	redisClient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", testRedisHost, testRedisPort)})
	redisClient.Set(testRedisKey, []byte(testData), 0)

	stateIface, err := NewStateFromRedis(testRedisHost, testRedisPort, testRedisKey)
	if err != nil {
		t.FailNow()
	}
	state := stateIface.(*redisState)

	if state.key != testRedisKey {
		t.Fail()
	}

	if state.AccessToken != "abc" {
		t.Fail()
	}

	redisClient.Del(testRedisKey)
}

func TestNewStateFromRedisMissingKey(t *testing.T) {
	state, err := NewStateFromRedis(testRedisHost, testRedisPort, testRedisKey)
	if err == nil {
		t.Fail()
	}

	if state != nil {
		t.Fail()
	}
}

func TestNewStateFromRedisCorruptKey(t *testing.T) {
	testData := `"garbage`
	redisClient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", testRedisHost, testRedisPort)})
	redisClient.Set(testRedisKey, []byte(testData), 0)

	state, err := NewStateFromRedis(testRedisHost, testRedisPort, testRedisKey)
	if err == nil {
		t.Fail()
	}

	if state != nil {
		t.Fail()
	}

	redisClient.Del(testRedisKey)
}

func TestRedisStateFirstUpdate(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey)
	now := time.Now()
	state.Update("xxx", "test", now)
	updateTime := state.GetLastUpdateTime("xxx", "test")
	if !updateTime.Equal(now) {
		t.Fail()
	}
}

func TestRedisStateSecondUpdate(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey)
	now := time.Now()
	state.Update("xxx", "test", now)

	prev := now.Add(-5 * time.Minute)
	state.Update("xxx", "test", prev)

	updateTime := state.GetLastUpdateTime("xxx", "test")
	if updateTime.Equal(now) {
		t.Fail()
	}
}

func TestRedisStateGetLastUptimeTimeNil(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey)
	lastUpdate := state.GetLastUpdateTime("xxx", "test")
	if !lastUpdate.Equal(time.Time{}) {
		t.Fail()
	}
}

func TestRedisStateAccessToken(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey)
	state.SetAccessToken("xxx")
	if state.GetAccessToken() != "xxx" {
		t.Fail()
	}
}

func TestRedisStateSave(t *testing.T) {
	state := NewRedisState(testRedisHost, testRedisPort, testRedisKey)
	state.SetAccessToken("xxx")
	state.Save()

	stateIface, err := NewStateFromRedis(testRedisHost, testRedisPort, testRedisKey)
	if err != nil {
		t.FailNow()
	}
	stateFromRedis := stateIface.(*redisState)
	if stateFromRedis.GetAccessToken() != "xxx" {
		t.Fail()
	}

	os.Remove("test.json")
}
