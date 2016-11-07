package state

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type State interface {
	GetAccessToken() string
	SetAccessToken(string)

	Save() error
	Update(string, string, time.Time)
	GetLastUpdateTime(string, string) time.Time
}

type fileState struct {
	AccessToken string
	Filename    string `json:"-"`
	// uuid -> reading_type -> timestamp
	LastUpdated map[string]map[string]time.Time
}

func NewFileState(filename string) State {
	return &fileState{
		Filename:    filename,
		LastUpdated: make(map[string]map[string]time.Time),
	}
}

// Tries to read state from file
func NewStateFromFile(filename string) (State, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	state := new(fileState)
	err = json.Unmarshal(data, state)
	if err != nil {
		return nil, err
	}
	state.Filename = filename

	return state, nil
}

// Overwrites the state file with the current state
func (s *fileState) Save() error {
	data, _ := json.Marshal(s)
	return ioutil.WriteFile(s.Filename, data, 0600)
}

// Helper func to make interacting with the LastUpdated map easierr
func (s *fileState) Update(uuid string, readingType string, timestamp time.Time) {
	if s.LastUpdated[uuid] == nil {
		s.LastUpdated[uuid] = make(map[string]time.Time)
	}
	s.LastUpdated[uuid][readingType] = timestamp
}

func (s *fileState) GetLastUpdateTime(uuid string, queryType string) time.Time {
	if s.LastUpdated[uuid] == nil {
		return time.Time{}
	}
	return s.LastUpdated[uuid][queryType]
}

func (s *fileState) GetAccessToken() string {
	return s.AccessToken
}

func (s *fileState) SetAccessToken(token string) {
	s.AccessToken = token
}
