package util

import (
	"os"
	"strings"
)

func Getenv(key, v string) string {
	if s := os.Getenv(key); strings.TrimSpace(s) != "" {
		return s
	}
	return v
}
