package main

import (
	"testing"
	"time"

	"github.com/arcticfoxnv/oolong/wirelesstag"
)

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
