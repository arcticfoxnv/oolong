package tsdb

import (
	"github.com/arcticfoxnv/oolong/wirelesstag"
)

// TSDB provides an interface for storing readings in a time series database.
type TSDB interface {
	PutValue(*wirelesstag.Tag, string, wirelesstag.Reading) error
}
