package main

import (
	"testing"
	"time"

	"github.com/arcticfoxnv/oolong/wirelesstag"
)

func TestNewOpenTSDBClient(t *testing.T) {
	c := NewOpenTSDBClient("localhost", 12345, "test")
	if c.client == nil {
		t.Fail()
	}
	if c.prefix != "test" {
		t.Fail()
	}
}

func TestPrepareValue(t *testing.T) {
	c := NewOpenTSDBClient("localhost", 12345, "test")

	tag := &wirelesstag.Tag{
		Name: "tag1",
		UUID: "xxx-yyy-zzz",
	}
	valueType := "widget"
	reading := wirelesstag.Reading{
		Timestamp: time.Now(),
		Value:     10.0,
	}

	data := c.prepareValue(tag, valueType, reading)
	if data.Timestamp != reading.Timestamp.Unix() {
		t.Fail()
	}

	if data.Value != reading.Value {
		t.Fail()
	}

	if data.Tags["uuid"] != tag.UUID {
		t.Fail()
	}

	if data.Tags["name"] != tag.Name {
		t.Fail()
	}

	if data.Metric != "test.widget" {
		t.Fail()
	}

}
