package utils

import (
	"strconv"
	"strings"
)

// ParseInt64List parses a comma-separated list of int64 values.
func ParseInt64List(raw string) []int64 {
	var result []int64
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

// ParseStringList parses a comma-separated list of strings, trimming whitespace.
func ParseStringList(raw string) []string {
	var result []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
