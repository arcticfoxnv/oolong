package main

import (
	"testing"
	"time"

	"github.com/arcticfoxnv/oolong/wirelesstag"
)

type DummyTagClient struct {
	Stats []wirelesstag.RawMultiStat
}

func (c *DummyTagClient) GetTagManagerTagList() (map[string][]wirelesstag.Tag, error) {
	tags := make(map[string][]wirelesstag.Tag)
	tags["abc"] = []wirelesstag.Tag{
		{
			Name: "tag1",
		},
		{
			Name: "tag2",
		},
	}
	return tags, nil
}

func (c *DummyTagClient) GetMultiTagStatsRaw([]int, string, time.Time, time.Time) ([]wirelesstag.RawMultiStat, error) {
	return c.Stats, nil
}

func (c *DummyTagClient) GetStatsRaw(int, time.Time, time.Time) ([]wirelesstag.RawStat, error) {
	return nil, nil
}

func (c *DummyTagClient) GetTagManagers() ([]wirelesstag.TagManager, error) {
	return nil, nil
}

var conversionTests = [][]float32{
	[]float32{-40, -40},
	[]float32{0, 32},
	[]float32{100, 212},
	[]float32{25, 77},
	[]float32{22, 71.6},
}

func TestConvertCToF(t *testing.T) {
	for _, test := range conversionTests {
		if ConvertCToF(test[0]) != test[1] {
			t.Fail()
		}
	}
}

func TestConvertFToC(t *testing.T) {
	for _, test := range conversionTests {
		if ConvertFToC(test[1]) != test[0] {
			t.Fail()
		}
	}
}

func TestFilterNewStatsNoFiltered(t *testing.T) {
	now := time.Now()
	stat := wirelesstag.Stat{
		SlaveId: 0,
		Readings: []wirelesstag.Reading{
			{
				Timestamp: now,
			},
			{
				Timestamp: now.Add(-1 * time.Minute),
			},
		},
	}
	filtered := FilterNewStats(stat, time.Time{})
	if len(filtered.Readings) != len(stat.Readings) {
		t.Fail()
	}
}

func TestFilterNewStatsAllFiltered(t *testing.T) {
	now := time.Now()
	stat := wirelesstag.Stat{
		SlaveId: 0,
		Readings: []wirelesstag.Reading{
			{
				Timestamp: now,
			},
			{
				Timestamp: now.Add(-1 * time.Minute),
			},
		},
	}
	filtered := FilterNewStats(stat, now.Add(time.Second))
	if len(filtered.Readings) != 0 {
		t.Fail()
	}
}

func TestFilterNewStatsSomeFiltered(t *testing.T) {
	now := time.Now()
	stat := wirelesstag.Stat{
		SlaveId: 0,
		Readings: []wirelesstag.Reading{
			{
				Timestamp: now,
			},
			{
				Timestamp: now.Add(-1 * time.Minute),
			},
		},
	}
	filtered := FilterNewStats(stat, now.Add(-1*time.Second))
	if len(filtered.Readings) != 1 {
		t.Fail()
	}
}

func TestGetTagBySlaveId(t *testing.T) {
	tags := []wirelesstag.Tag{
		{
			SlaveId: 1,
			Name:    "test1",
		},
		{
			SlaveId: 0,
			Name:    "test4",
		},
	}
	tag := GetTagBySlaveId(tags, 0)
	if tag.SlaveId != 0 {
		t.Fail()
	}
}

func TestGetTagBySlaveIdBadTagId(t *testing.T) {
	tags := []wirelesstag.Tag{
		{
			SlaveId: 1,
			Name:    "test1",
		},
		{
			SlaveId: 0,
			Name:    "test4",
		},
	}
	tag := GetTagBySlaveId(tags, 2)
	if tag != nil {
		t.Fail()
	}
}

func TestGetTagBySlaveIdNilTags(t *testing.T) {
	tag := GetTagBySlaveId(nil, 2)
	if tag != nil {
		t.Fail()
	}
}

func TestGetTags(t *testing.T) {
	client := &DummyTagClient{}
	tags, err := GetTags(client)

	if err != nil {
		t.Fail()
	}

	if len(tags) != 2 {
		t.Fail()
	}
}

func TestGetStats(t *testing.T) {
	client := &DummyTagClient{
		Stats: []wirelesstag.RawMultiStat{
			{
				Date: "1/2/2006",
				SlaveIds: []int{
					0,
					1,
				},
				Values: [][]float32{
					{
						1,
						2,
					},
					{
						3,
						4,
					},
				},
				TimeOfDaySeconds: [][]int{
					{
						0,
						5,
					},
					{
						1,
						605,
					},
				},
			},
		},
	}

	stats, err := GetStats(client, "whatever", []int{0, 1}, time.Now(), time.Now())
	if err != nil {
		t.Fail()
	}

	if len(stats) != 2 {
		t.Fail()
	}

	if stats[0].Readings[0].Value != 1 {
		t.Fail()
	}
}
