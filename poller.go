package main

import (
	"log"
	"time"

	"github.com/arcticfoxnv/oolong/tsdb"
	"github.com/arcticfoxnv/oolong/wirelesstag"
)

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
	lastFetchTime := time.Now()

	for {
		startDay := time.Now()
		endDay := startDay

		// Check if we started a new day.  If so, we need to include yesterday
		// in the next fetch to grab any readings added between the last fetch
		// and end of day.
		if startDay.Hour() < lastFetchTime.Hour() {
			startDay = startDay.Add(-24 * time.Hour)
			log.Printf("New day started.  Adjusting query to include end of day %s", startDay.Format("2006-01-02"))
		}

		// We need to do a separate query for each of the stats we want.
		// The GetStatsRaw API method gets all of these (except battery it seems)
		// in one call, but only for a single tag.  The GetMultiTagStatsRaw API
		// method only returns one stat, but for multiple tags.
		// We're using the Multi method, which should save on total API calls
		// once the number of tags we query per call is more than 3.
		for _, queryType := range config.QueryStats {
			stats, err := GetStats(tagClient, queryType, tagIds, startDay, endDay)
			if err != nil {
				log.Printf("Failed to load raw %s stats: %s\n", queryType, err.Error())
				continue
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
		lastFetchTime = time.Now()

		// Sleep to avoid excessive calls to the API server.
		time.Sleep(time.Duration(config.PollInterval) * time.Second)
	}
}

func GetTags(tagClient wirelesstag.Client) ([]wirelesstag.Tag, error) {
	// We could probably call GetTagList instead, and simply this function,
	// but we might want to add support later on for tracking which tags are
	// connected to which tag managers.
	tagManagersAndTags, err := tagClient.GetTagManagerTagList()
	if err != nil {
		return nil, err
	}

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

func GetStats(tagClient wirelesstag.Client, queryType string, ids []int, start, end time.Time) ([]wirelesstag.Stat, error) {
	// Query the API server for the specified stat (temp, humidity, battery, etc) and tags.
	rawStats, err := tagClient.GetMultiTagStatsRaw(ids, queryType, start, end)
	if err != nil {
		return nil, err
	}

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
