package version

import (
	"strconv"
	"time"
)

// Build time variables with default values
// These are set at compile time
var (
	version   string
	timestamp string
	commit    string
)

// Version returns the build version
func Version() string {
	return version
}

// Commit returns the build commit hash
func Commit() string {
	return commit
}

// BuildTimestamp returns the build timestamp
func BuildTime() time.Time {
	ts, _ := strconv.ParseInt(timestamp, 10, 64)
	return time.Unix(ts, 0).UTC()
}
