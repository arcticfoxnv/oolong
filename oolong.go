package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "time"

    "github.com/arcticfoxnv/oolong/oauth"
    "github.com/arcticfoxnv/oolong/tsdb"
    "github.com/arcticfoxnv/oolong/wirelesstag"
    "github.com/BurntSushi/toml"
)

type Config struct {
    OAuth OAuthConfig
    HTTP HTTPConfig
    PollInterval int `toml:"poll_interval"`
    QueryStats []string `toml:"query_stats"`
    ConvertToF bool `toml:"convert_to_f"`
    OpenTSDB OpenTSDBConfig
}

type HTTPConfig struct {
    Port int
}

// These are found at https://mytaglist.com/eth/oauth2_apps.html
type OAuthConfig struct {
    ID string
    Secret string
}

type OpenTSDBConfig struct {
    Host string
    Port int
    MetricsPrefix string `toml:"metrics_prefix"`
}

type OolongState struct {
    AccessToken string
    // uuid -> reading_type -> timestamp
    LastUpdated map[string]map[string]time.Time
}

var tagClient wirelesstag.Client
var clientReady chan int
var oolongState *OolongState

const oolongStateFile = "state.json"

func ReadConfigFile(filename string) *Config {
    config := new(Config)
    tomlData, err := ioutil.ReadFile(filename)
    if err != nil {
        log.Fatalf("Failed to read config file: %s\n", err.Error())
    }
    _, err = toml.Decode(string(tomlData), config)
    if err != nil {
        log.Fatalf("Failed to parse config file: %s\n", err.Error())
    }
    return config
}

func NewOolongState() *OolongState {
    return &OolongState{
        LastUpdated: make(map[string]map[string]time.Time),
    }
}

// Tries to read state from file
func ReadOolongState() (*OolongState, error) {
    data, err := ioutil.ReadFile(oolongStateFile)
    if err != nil { return nil, err }
    state := new(OolongState)
    err = json.Unmarshal(data, state)
    if err != nil { return nil, err }

    return state, nil
}

// Overwrites the state file with the current state
func (s *OolongState) WriteToFile() (error) {
    data, _ := json.Marshal(oolongState)
    return ioutil.WriteFile(oolongStateFile, data, 0600)
}

// Helper func to make interacting with the LastUpdated map easierr
func (s *OolongState) UpdateLastUpdatedTime(uuid string, readingType string, timestamp time.Time) {
    if oolongState.LastUpdated[uuid] == nil {
        oolongState.LastUpdated[uuid] = make(map[string]time.Time)
    }
    oolongState.LastUpdated[uuid][readingType] = timestamp
}

func StatsFetcher(config *Config, tsdbClient tsdb.TSDB) {
    // Wait for Wireless Tag API client to be ready
    <- clientReady

    // Get tag list
    log.Printf("Fetching list of tags...\n")
    tags, err := GetTags()
    if err != nil {
        log.Fatalf("Failed to load tags: %s\n", err.Error())
    }

    // Extract ids for use later
    var tagIds []int
    for _, t := range tags {
        tagIds = append(tagIds, t.SlaveId)
    }

    for {
        // We need to do a separate query for each of the stats we want.
        // The GetStatsRaw API method gets all of these (except battery it seems)
        // in one call, but only for a single tag.  The GetMultiTagStatsRaw API
        // method only returns one stat, but for multiple tags.
        // We're using the Multi method, which should save on total API calls
        // once the number of tags we query per call is more than 3.
        for _, queryType := range config.QueryStats {
            stats, err := GetStats(queryType, tagIds)
            if err != nil {
                // TODO: Maybe this shouldn't be fatal anymore?  Now that we record
                // the last successful reading, the next successful call will read
                // anything a failed call should have.
                log.Fatalf("Failed to load raw %s stats: %s\n", queryType, err.Error())
            }
            log.Printf("Fetched %s stats for %d tags\n", queryType, len(stats))
            // Iterate through each returned stat (one stat per tag)
            for _, stat := range stats {
                // Stats return tags by SlaveId, but we store tags in state/datastore
                // by UUID.
                tag := GetTagBySlaveId(tags, stat.SlaveId)

                // Determine the last time this stat for this tag was updated
                lastUpdated := time.Time{}
                if oolongState.LastUpdated[tag.UUID] != nil {
                    lastUpdated = oolongState.LastUpdated[tag.UUID][queryType]
                }

                // Filter out old readings
                newStat := FilterNewStats(stat, lastUpdated)
                log.Printf("  * Fetched %d new %s stats for tag %s (%d)", len(newStat.Readings), queryType, tag.UUID, stat.SlaveId)

                // Process all of the remaining readings for this stat/tag
                for _, reading := range newStat.Readings {

                    // Special handling for temperature - values are returned from
                    // the API in celsius, but for those unlucky few who grew up
                    // learning fahrenheit instead, convert to something we can read
                    if queryType == "temperature" && config.ConvertToF {
                        reading.Value = ConvertCToF(reading.Value)
                    }

                    // Store value in the data store
                    err := tsdbClient.PutValue(tag, queryType, reading)
                    if err != nil {
                        log.Fatalf("Failed to store value: %s\n", err.Error())
                    }
                    // Update the state with new timestamp
                    oolongState.UpdateLastUpdatedTime(tag.UUID, queryType, reading.Timestamp)
                }
            }
        }

        // Once all of the stats have been processed, update the state file on disk
        oolongState.WriteToFile()

        // Sleep to avoid excessive calls to the API server.
        time.Sleep(time.Duration(config.PollInterval) * time.Second)
    }
}

func GetTags() ([]wirelesstag.Tag, error) {
    // We could probably call GetTagList instead, and simply this function,
    // but we might want to add support later on for tracking which tags are
    // connected to which tag managers.
    tagManagersAndTags, err := tagClient.GetTagManagerTagList()
    if err != nil { return nil, err }

    tagList := []wirelesstag.Tag{}
    for _, tags := range tagManagersAndTags {
        for _, tag := range tags {
            tagList = append(tagList, tag)
        }
    }
    return tagList, nil
}

func GetTagBySlaveId(tags []wirelesstag.Tag, slaveId int) *wirelesstag.Tag {
    for _, t := range tags {
        if t.SlaveId == slaveId {
            return &t
        }
    }
    return nil
}

func GetStats(queryType string, ids []int) ([]wirelesstag.Stat, error) {
    // Query the API server for the specified stat (temp, humidity, battery, etc) and tags.
    // The API supports a start and end date, but for a continuously running background process,
    // we only care about today.
    rawStats, err := tagClient.GetMultiTagStatsRaw(ids, queryType, time.Now(), time.Now())
    if err != nil { return nil, err }

    // The results are returned in a set of nested arrays, with timestamps as offsets
    // starting from midnight of the requested day.
    // Let's transform that into something easier to handle.
    return wirelesstag.NormalizeRawMultiStat(rawStats), nil
}

func FilterNewStats(stat wirelesstag.Stat, lastReadTime time.Time) wirelesstag.Stat {
    newStats := wirelesstag.Stat{SlaveId: stat.SlaveId}
    for _, reading := range stat.Readings {
        if reading.Timestamp.After(lastReadTime) {
            newStats.Readings = append(newStats.Readings, reading)
        }
    }
    return newStats
}

// Convert celsius to fahrenheit for the poor bastards who grew up in a
// non-metric system country
func ConvertCToF(temp float32) float32 {
    return (temp * 9 / 5) + 32
}

// Might as well have a function to convert back.
func ConvertFToC(temp float32) float32 {
    return (temp - 32) * 5 / 9
}

func main() {
    // Read config file
    config := ReadConfigFile("oolong.toml")

    // Initialize data storage client
    tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

    // Channel to signal the wireless tag client is ready to read data
    clientReady = make(chan int, 1)

    // Try to load state from file
    oolongState, err = ReadOolongState()
    if err != nil {
        log.Printf("Error: %s\n", err.Error())
        log.Printf("Unable to restore state.  Starting HTTP server...\n")
        startChan := make(chan string)

        // HTTP server is only needed to do OAuth fun.  The resulting token
        // is saved to the state file for reuse.
        go StartHTTPServer(&config, startChan)

        // Wait for the server to initialize
        startUrl := <- startChan

        // Print the URL for the user to visit to start the OAuth dance.
        log.Printf("Oolong ready!  Go to %s to begin.", startUrl)
    } else {

        // Use token from state file to initialize the wireless tag client
        tagClient = wirelesstag.NewClient(oolongState.AccessToken)
        clientReady <- 1
        log.Printf("Oolong ready!\n")
    }

    // Retrieve stats from cloud and push to data storage
    StatsFetcher(&config, tsdbClient)
}
