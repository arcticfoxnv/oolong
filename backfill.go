package main

import (
	"log"
	"time"

	"github.com/arcticfoxnv/oolong/tsdb"
	"github.com/arcticfoxnv/oolong/wirelesstag"
)

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
		stats, err := GetStats(tagClient, queryType, tagIds, date, date)
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
