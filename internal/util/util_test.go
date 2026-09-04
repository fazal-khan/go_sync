package util

import (
	"os"
	"testing"
)

func TestSubstituteEnvVars_BasicVar(t *testing.T) {
	t.Setenv("NAME", "world")
	got := SubstituteEnvVars("hello ${NAME}")
	want := "hello world"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_EnvPrefix(t *testing.T) {
	t.Setenv("test_sync.USER", "admin")
	got := SubstituteEnvVars("user=${env.test_sync.USER}")
	want := "user=admin"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_RowUntouched(t *testing.T) {
	got := SubstituteEnvVars(`"id": "${row.id}"`)
	want := `"id": "${row.id}"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_MissingVar(t *testing.T) {
	os.Unsetenv("DOES_NOT_EXIST_98765")
	got := SubstituteEnvVars("${DOES_NOT_EXIST_98765}")
	want := ""
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_NoVars(t *testing.T) {
	got := SubstituteEnvVars("plain text without vars")
	want := "plain text without vars"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_MultipleVars(t *testing.T) {
	t.Setenv("A", "x")
	t.Setenv("B", "y")
	got := SubstituteEnvVars("${A}-${B}")
	want := "x-y"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_MixedRowAndEnv(t *testing.T) {
	t.Setenv("USER", "admin")
	got := SubstituteEnvVars("${env.USER} - ${row.id}")
	want := "admin - ${row.id}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_EmptyString(t *testing.T) {
	got := SubstituteEnvVars("")
	want := ""
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSubstituteEnvVars_EnvPrefixEmptyKey(t *testing.T) {
	// ${env.} with no actual key after prefix → looks up empty string ""
	os.Unsetenv("")
	got := SubstituteEnvVars("prefix-${env.}suffix")
	want := "prefix-suffix"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetenv_WithEnv(t *testing.T) {
	t.Setenv("MY_KEY", "my_value")
	got := Getenv("MY_KEY", "default")
	want := "my_value"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetenv_Missing(t *testing.T) {
	os.Unsetenv("MISSING_KEY_XYZ")
	got := Getenv("MISSING_KEY_XYZ", "fallback")
	want := "fallback"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetenv_EmptyString(t *testing.T) {
	t.Setenv("EMPTY_KEY", "")
	got := Getenv("EMPTY_KEY", "fallback")
	want := "fallback"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseNonNegativeInt(t *testing.T) {
	cases := map[string]int{
		"100":   100,
		"  50 ": 50,
		"1":     1,
		"200":   200,
		"0":     0,
		"-5":    0,
		"-3":    0,
		"-1":    0,
		"abc":   0,
		"x":     0,
		"nope":  0,
		"":      0,
	}
	for in, want := range cases {
		if got := ParseNonNegativeInt(in); got != want {
			t.Errorf("ParseNonNegativeInt(%q) = %d, want %d", in, got, want)
		}
	}
}
