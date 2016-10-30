package main

import (
	"fmt"
	"strings"

	"github.com/arcticfoxnv/oolong/wirelesstag"
	"github.com/bluebreezecf/opentsdb-goclient/client"
	"github.com/bluebreezecf/opentsdb-goclient/config"
)

type OpenTSDB struct {
	client client.Client
	prefix string
}

func NewOpenTSDBClient(host string, port int, metricPrefix string) *OpenTSDB {
	cfg := config.OpenTSDBConfig{OpentsdbHost: fmt.Sprintf("%s:%d", host, port)}
	c, _ := client.NewClient(cfg)
	return &OpenTSDB{
		client: c,
		prefix: metricPrefix,
	}
}

func (c *OpenTSDB) PutValue(tag *wirelesstag.Tag, valueType string, reading wirelesstag.Reading) error {
	// Put the data into the new structure for submitting to opentsdb
	data := client.DataPoint{
		Metric:    fmt.Sprintf("%s.%s", c.prefix, valueType),
		Timestamp: reading.Timestamp.Unix(),
		Value:     reading.Value,
		Tags:      make(map[string]string),
	}
	// For now, tag with both UUID and Name.  We can use these to filter/display
	// on dashboards
	data.Tags["uuid"] = tag.UUID
	data.Tags["name"] = strings.Replace(tag.Name, " ", "_", -1)

	// Submit value to opentsdb
	_, err := c.client.Put([]client.DataPoint{data}, "summary")
	return err
}
