package utils

import "time"

func SelectTime(unit string, timeoutValue int) time.Duration {
	switch unit {
	case "seconds":
		return time.Duration(timeoutValue) * time.Second
	case "minutes":
		return time.Duration(timeoutValue) * time.Minute
	case "hours":
		return time.Duration(timeoutValue) * time.Hour
	default:
		return time.Duration(timeoutValue) * time.Second
	}
}
