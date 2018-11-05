package timeutil

import (
	"fmt"
	"time"
)

func GetTimestamp(time *time.Time) string {
	return fmt.Sprintf("%v-%v-%02vT%02v:%02v:%02v+00:00", time.Year(), int(time.Month()), time.Day(), time.Hour(), time.Minute(), time.Second())
}

func GetTimeDifferenceInMilliseconds(startTime, endTime int64) (ms int64) {
	return (endTime - startTime) / 1000 / 1000 //convert ns to ms
}
