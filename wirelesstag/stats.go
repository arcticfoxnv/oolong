package wirelesstag

import (
	"time"
)

const DateFormat = "1/2/2006"

type RawStat struct {
	Date             string
	Temperatures     []float32 `json:"temps"`
	Caps             []float32
	TimeOfDaySeconds []int `json:"tods"`
}

type RawMultiStat struct {
	Date             string
	SlaveIds         []int `json:"ids"`
	Values           [][]float32
	TimeOfDaySeconds [][]int `json:"tods"`
}

type Stat struct {
	SlaveId  int
	Readings []Reading
}

type Reading struct {
	Timestamp time.Time
	Value     float32
}

func NormalizeRawMultiStat(rawStats []RawMultiStat) ([]Stat, error) {
	normalizedStats := []Stat{}

	// Maps slave_id -> index+1 in normStats
	deviceMap := make(map[int]int, 0)

	for _, dayStat := range rawStats {
		date, err := time.ParseInLocation(DateFormat, dayStat.Date, time.Local)
		if err != nil {
			return nil, err
		}

		// Allocate space in normStats for each device
		for _, slaveId := range dayStat.SlaveIds {

			// Maps return the default if the entry isn't found.
			if deviceMap[slaveId] == 0 {
				// Append an empty Stat to the array
				normalizedStats = append(normalizedStats, Stat{SlaveId: slaveId})

				// Normally, this should be len-1, but since maps return 0 for not found,
				// whatever device was assigned to index 0 would have a second entry created.
				deviceMap[slaveId] = len(normalizedStats)
			}
		}

		for deviceValueIndex, deviceValues := range dayStat.Values {
			normalizedStatsIndex := deviceMap[dayStat.SlaveIds[deviceValueIndex]] - 1
			for valueIndex, value := range deviceValues {
				timestamp := date.Add(time.Duration(dayStat.TimeOfDaySeconds[deviceValueIndex][valueIndex]) * time.Second)
				normalizedReading := Reading{
					Timestamp: timestamp,
					Value:     value,
				}
				normalizedStats[normalizedStatsIndex].Readings = append(normalizedStats[normalizedStatsIndex].Readings, normalizedReading)
			}
		}
	}

	return normalizedStats, nil
}
