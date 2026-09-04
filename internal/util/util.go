package util

import (
	"fmt"
	"log/slog"
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
	val, err := substituteEnvVars(s, map[string]bool{})
	if err != nil {
		slog.Default().Error(err.Error())
	}
	return val
}

func substituteEnvVars(s string, resolving map[string]bool) (string, error) {
	var err error

	s = envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		if err != nil {
			return match
		}

		key := match[2 : len(match)-1]

		// Leave ${row.*} untouched.
		if strings.HasPrefix(key, "row.") {
			return match
		}

		envKey := strings.TrimPrefix(key, "env.")

		// Detect cycles.
		if resolving[envKey] {
			err = fmt.Errorf("cyclic environment variable reference: %s", envKey)
			return match
		}

		value := os.Getenv(envKey)

		// Resolve placeholders inside the env value.
		resolving[envKey] = true
		value, err = substituteEnvVars(value, resolving)
		delete(resolving, envKey)

		return value
	})

	return s, err
}

// ParseNonNegativeInt parses s as a non-negative integer, returning 0 for
// empty, non-numeric, or negative input.
func ParseNonNegativeInt(s string) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return 0
}
