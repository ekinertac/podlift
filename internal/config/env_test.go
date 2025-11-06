package config

import (
	"os"
	"strings"
	"testing"
)

func TestSubstituteEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envKey   string
		envValue string
		want     string
	}{
		{
			name:     "simple substitution",
			input:    "${TEST_VAR}",
			envKey:   "TEST_VAR",
			envValue: "test_value",
			want:     "test_value",
		},
		{
			name:     "with default used",
			input:    "${MISSING_VAR:-default}",
			envKey:   "",
			envValue: "",
			want:     "default",
		},
		{
			name:     "with default not used",
			input:    "${TEST_VAR:-default}",
			envKey:   "TEST_VAR",
			envValue: "actual",
			want:     "actual",
		},
		{
			name:     "in middle of string",
			input:    "prefix-${TEST_VAR}-suffix",
			envKey:   "TEST_VAR",
			envValue: "middle",
			want:     "prefix-middle-suffix",
		},
		{
			name:     "multiple vars",
			input:    "${VAR1}-${VAR2}",
			envKey:   "VAR1",
			envValue: "value1",
			want:     "value1-${VAR2}", // VAR2 not set
		},
		{
			name:     "no substitution",
			input:    "plain text",
			envKey:   "",
			envValue: "",
			want:     "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			got := SubstituteEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("SubstituteEnvVars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(string) bool
	}{
		{
			name:  "tilde expansion",
			input: "~/.ssh/id_rsa",
			check: func(expanded string) bool {
				return len(expanded) > 0 && expanded[0] != '~'
			},
		},
		{
			name:  "absolute path unchanged",
			input: "/absolute/path",
			check: func(expanded string) bool {
				return expanded == "/absolute/path"
			},
		},
		{
			name:  "relative path unchanged",
			input: "relative/path",
			check: func(expanded string) bool {
				return expanded == "relative/path"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if !tt.check(got) {
				t.Errorf("ExpandPath(%q) = %q, check failed", tt.input, got)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	// Create temp .env file
	tmpfile, err := os.CreateTemp("", ".env-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `# Comment
KEY1=value1
KEY2="value2"
KEY3='value3'
EMPTY=
# Another comment
KEY4=value4
`
	tmpfile.WriteString(content)
	tmpfile.Close()

	// Load env
	err = LoadEnv(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadEnv() error = %v", err)
	}

	// Check values
	if os.Getenv("KEY1") != "value1" {
		t.Errorf("KEY1 = %v, want value1", os.Getenv("KEY1"))
	}

	if os.Getenv("KEY2") != "value2" {
		t.Errorf("KEY2 = %v, want value2 (quotes removed)", os.Getenv("KEY2"))
	}

	if os.Getenv("KEY3") != "value3" {
		t.Errorf("KEY3 = %v, want value3 (quotes removed)", os.Getenv("KEY3"))
	}

	if os.Getenv("KEY4") != "value4" {
		t.Errorf("KEY4 = %v, want value4", os.Getenv("KEY4"))
	}

	// Cleanup
	os.Unsetenv("KEY1")
	os.Unsetenv("KEY2")
	os.Unsetenv("KEY3")
	os.Unsetenv("KEY4")
}

func TestLoadEnv_NonExistent(t *testing.T) {
	// Should not error if file doesn't exist
	err := LoadEnv("/nonexistent/path/.env")
	if err != nil {
		t.Errorf("LoadEnv() should not error for non-existent file, got: %v", err)
	}
}

func TestFindEnvFile(t *testing.T) {
	// Create temp directory with .env
	tmpdir, err := os.MkdirTemp("", "podlift-env-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	envPath := tmpdir + "/.env"
	os.WriteFile(envPath, []byte("TEST=value"), 0644)

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(oldDir)

	// Find should find it
	found := FindEnvFile()
	if found == "" {
		t.Error("FindEnvFile() should find .env file")
	}
	if !strings.HasSuffix(found, ".env") {
		t.Errorf("FindEnvFile() = %v, should end with .env", found)
	}
}

