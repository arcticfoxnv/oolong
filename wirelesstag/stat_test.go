package wirelesstag

import (
	"testing"
	"time"
)

func TestNormalizeRawMultiStat(t *testing.T) {
	raw := []RawMultiStat{
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
	}

	output := NormalizeRawMultiStat(raw)
	if output[0].SlaveId != 0 {
		t.Fail()
	}

	if len(output[0].Readings) != 2 {
		t.Fail()
	}

	expectedTimestamp := time.Date(2006, 1, 2, 0, 0, 0, 0, time.Local)
	if !output[0].Readings[0].Timestamp.Equal(expectedTimestamp) {
		t.Fail()
	}
	if output[0].Readings[0].Value != 1 {
		t.Fail()
	}

	expectedTimestamp = time.Date(2006, 1, 2, 0, 10, 5, 0, time.Local)
	if !output[1].Readings[1].Timestamp.Equal(expectedTimestamp) {
		t.Fail()
	}
	if output[1].Readings[1].Value != 4 {
		t.Fail()
	}
}
