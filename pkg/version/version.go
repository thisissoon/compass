package version

import (
	"fmt"
	"strconv"
	"time"
)

var (
	Version      string
	Timestamp    string
	GitCommit    string
	GitTreeState string
)

// BuildTimestamp returns the build timestamp
func BuildTime() time.Time {
	ts, _ := strconv.ParseInt(Timestamp, 10, 64)
	return time.Unix(ts, 0).UTC()
}

// Print version info
func Print() {
	fmt.Println("Version:", Version)
	fmt.Println("Build Time:", BuildTime().Format(time.RFC1123))
	fmt.Println("Git Commit:", GitCommit)
	fmt.Println("Git Tree State:", GitTreeState)
}
