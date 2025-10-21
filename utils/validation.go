package utils

import "regexp"

var customIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func IsValidCustomID(id string) bool {
	if len(id) > 64 || len(id) < 1 {
		return false
	}
	return customIDPattern.MatchString(id)
}
