package state

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/redis.v5"
)

type redisState struct {
	AccessToken string
	client      *redis.Client `json:"-"`
	key         string        `json:"-"`
	// uuid -> reading_type -> timestamp
	LastUpdated map[string]map[string]time.Time
}

func NewRedisState(host string, port int, key string) State {
	return &redisState{
		client:      redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", host, port)}),
		key:         key,
		LastUpdated: make(map[string]map[string]time.Time),
	}
}

func NewStateFromRedis(host string, port int, key string) (State, error) {
	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", host, port)})
	data, err := client.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	state := new(redisState)
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	state.client = client
	state.key = key

	return state, nil
}

func (s *redisState) Save() error {
	data, _ := json.Marshal(s)
	return s.client.Set(s.key, data, 0).Err()
}

// Helper func to make interacting with the LastUpdated map easierr
func (s *redisState) Update(uuid string, readingType string, timestamp time.Time) {
	if s.LastUpdated[uuid] == nil {
		s.LastUpdated[uuid] = make(map[string]time.Time)
	}
	s.LastUpdated[uuid][readingType] = timestamp
}

func (s *redisState) GetLastUpdateTime(uuid string, queryType string) time.Time {
	if s.LastUpdated[uuid] == nil {
		return time.Time{}
	}
	return s.LastUpdated[uuid][queryType]
}

func (s *redisState) GetAccessToken() string {
	return s.AccessToken
}

func (s *redisState) SetAccessToken(token string) {
	s.AccessToken = token
}
