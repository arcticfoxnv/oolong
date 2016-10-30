package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "time"

    "github.com/arcticfoxnv/oolong/tsdb"
    "github.com/arcticfoxnv/oolong/wirelesstag"
    "github.com/BurntSushi/toml"
    "github.com/urfave/cli"
)

const Version = "0.0.1"

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
    data, _ := json.Marshal(s)
    return ioutil.WriteFile(oolongStateFile, data, 0600)
}

// Helper func to make interacting with the LastUpdated map easierr
func (s *OolongState) UpdateLastUpdatedTime(uuid string, readingType string, timestamp time.Time) {
    if s.LastUpdated[uuid] == nil {
        s.LastUpdated[uuid] = make(map[string]time.Time)
    }
    s.LastUpdated[uuid][readingType] = timestamp
}

func StatsFetcher(config *Config, state *OolongState, tagClient wirelesstag.Client, tsdbClient tsdb.TSDB) {

    // Get tag list
    log.Printf("Fetching list of tags...\n")
    tags, err := GetTags(tagClient)
    if err != nil {
        log.Fatalf("Failed to load tags: %s\n", err.Error())
    }

    // Extract ids for use later
    var tagIds []int
    for _, t := range tags {
        tagIds = append(tagIds, t.SlaveId)
    }

    for {
        today := time.Now()
        // We need to do a separate query for each of the stats we want.
        // The GetStatsRaw API method gets all of these (except battery it seems)
        // in one call, but only for a single tag.  The GetMultiTagStatsRaw API
        // method only returns one stat, but for multiple tags.
        // We're using the Multi method, which should save on total API calls
        // once the number of tags we query per call is more than 3.
        for _, queryType := range config.QueryStats {
            stats, err := GetStats(tagClient, queryType, tagIds, today)
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
                if state.LastUpdated[tag.UUID] != nil {
                    lastUpdated = state.LastUpdated[tag.UUID][queryType]
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
                    state.UpdateLastUpdatedTime(tag.UUID, queryType, reading.Timestamp)
                }
            }
        }

        // Once all of the stats have been processed, update the state file on disk
        state.WriteToFile()

        // Sleep to avoid excessive calls to the API server.
        time.Sleep(time.Duration(config.PollInterval) * time.Second)
    }
}

func Backfill(config *Config, state *OolongState, tagClient wirelesstag.Client, tsdbClient tsdb.TSDB, date time.Time) {

    // Get tag list
    log.Printf("Fetching list of tags...\n")
    tags, err := GetTags(tagClient)
    if err != nil {
        log.Fatalf("Failed to load tags: %s\n", err.Error())
    }

    // Extract ids for use later
    var tagIds []int
    for _, t := range tags {
        tagIds = append(tagIds, t.SlaveId)
    }

    for _, queryType := range config.QueryStats {
        stats, err := GetStats(tagClient, queryType, tagIds, date)
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

            log.Printf("  * Fetched %d %s stats for tag %s (%d)", len(stat.Readings), queryType, tag.UUID, stat.SlaveId)

            // Process all of the remaining readings for this stat/tag
            for _, reading := range stat.Readings {

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
            }
        }
    }
}

func GetTags(tagClient wirelesstag.Client) ([]wirelesstag.Tag, error) {
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

func GetStats(tagClient wirelesstag.Client, queryType string, ids []int, when time.Time) ([]wirelesstag.Stat, error) {
    // Query the API server for the specified stat (temp, humidity, battery, etc) and tags.
    // The API supports a start and end date, but for a continuously running background process,
    // we only care about today.
    rawStats, err := tagClient.GetMultiTagStatsRaw(ids, queryType, when, when)
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

func cmdHTTPServer(c *cli.Context) error {
    // Read config file
    config := ReadConfigFile(c.GlobalString("config"))

    // Channel to signal setup is complete.
    setupDone := make(chan int)
    startChan := make(chan string)

    // HTTP server is only needed to do OAuth fun.  The resulting token
    // is saved to the state file for reuse.
    log.Printf("Starting HTTP server...\n")
    go StartHTTPServer(config, startChan, setupDone)

    // Wait for the server to initialize
    startUrl := <- startChan

    // Print the URL for the user to visit to start the OAuth dance.
    log.Printf("Oolong ready!  Go to %s to begin.", startUrl)

    // Wait for setup to complete
    <- setupDone

    return nil
}

func cmdPoller(c *cli.Context) error {
    // Read config file
    config := ReadConfigFile(c.GlobalString("config"))

    // Initialize data storage client
    tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

    // Try to load state from file
    state, err := ReadOolongState()
    if err != nil {
        log.Fatalf("Unable to restore state: %s\n", err.Error())
    }

    // Use token from state file to initialize the wireless tag client
    tagClient := wirelesstag.NewClient(state.AccessToken)

    // Retrieve stats from cloud and push to data storage
    StatsFetcher(config, state, tagClient, tsdbClient)

    return nil
}

func cmdBackfill(c *cli.Context) error {
    if c.String("date") == "" {
        log.Fatalln("--date is required")
    }
    // Read config file
    config := ReadConfigFile(c.GlobalString("config"))
    date, err := time.Parse("2006-01-02", c.String("date"))
    if err != nil {
        log.Fatalf("Failed to parse date: %s\n", err.Error())
    }

    // Initialize data storage client
    tsdbClient := NewOpenTSDBClient(config.OpenTSDB.Host, config.OpenTSDB.Port, config.OpenTSDB.MetricsPrefix)

    // Try to load state from file
    state, err := ReadOolongState()
    if err != nil {
        log.Fatalf("Unable to restore state: %s\n", err.Error())
    }

    // Use token from state file to initialize the wireless tag client
    tagClient := wirelesstag.NewClient(state.AccessToken)

    // Retrieve stats from cloud and push to data storage
    Backfill(config, state, tagClient, tsdbClient, date)

    return nil
}

func main() {
    app := cli.NewApp()
    app.Name = "oolong"
    app.Usage = "Polls wirelesstag.net cloud data"
    app.Version = Version
    app.Flags = []cli.Flag {
        cli.StringFlag{
            Name: "config",
            Value: "oolong.toml",
            Usage: "Specifies path to config file",
        },
    }
    app.Commands = []cli.Command{
        {
            Name: "init",
            Usage: "Run the HTTP server to setup OAuth",
            Action: cmdHTTPServer,
        },
        {
            Name: "run",
            Usage: "Run the poller to periodically retrieve data from wirelesstag.net",
            Action: cmdPoller,
        },
        {
            Name: "backfill",
            Usage: "Retrieve a specific day and exit",
            Action: cmdBackfill,
            Flags: []cli.Flag{
                cli.StringFlag{
                    Name: "date",
                    Usage: "Specify the date to retrieve (format: YYYY-MM-DD)",
                },
            },
        },
    }

    app.Run(os.Args)
}
