package util

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Getenv(key, v string) string {
	if s := os.Getenv(key); strings.TrimSpace(s) != "" {
		return s
	}
	return v
}

// envVarPattern matches ${...} placeholders in a string.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// SubstituteEnvVars replaces ${...} patterns with environment variable values.
//   - ${env.XXX} → looks up os.Getenv("XXX") (removes the "env." prefix)
//   - ${VAR}     → looks up os.Getenv("VAR")
//   - ${row.*}   → left untouched (runtime row references)
//   - Missing env vars resolve to empty string.
func SubstituteEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the key inside ${...}
		key := match[2 : len(match)-1]

		// Leave ${row.*} untouched
		if strings.HasPrefix(key, "row.") {
			return match
		}

		// remove env. prefix if present
		envKey := strings.TrimPrefix(key, "env.")

		return os.Getenv(envKey)
	})
}

// ParseNonNegativeInt parses s as a non-negative integer, returning 0 for
// empty, non-numeric, or negative input.
func ParseNonNegativeInt(s string) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return 0
}
