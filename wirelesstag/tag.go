package wirelesstag

import ()

type Tag struct {
	Alive            bool
	BatteryRemaining float32
	BatteryVolt      float32
	Cap              float32
	Comment          string
	LastComm         int
	Name             string
	Rev              byte
	SlaveId          int
	TagType          int
	Temperature      float32
	UUID             string
	Version1         byte
}
