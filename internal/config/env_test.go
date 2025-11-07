package config

import (
	"os"
	"path/filepath"
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

	envPath := filepath.Join(tmpdir, ".env")
	os.WriteFile(envPath, []byte("TEST=value"), 0644)

	// Create config file path
	configPath := filepath.Join(tmpdir, "podlift.yml")

	// Find should find .env in same directory
	found := FindEnvFile(configPath)
	if found == "" {
		t.Error("FindEnvFile() should find .env file")
	}
	if found != envPath {
		t.Errorf("FindEnvFile() = %v, want %v", found, envPath)
	}
	
	// Should return empty string if config path is empty
	if found := FindEnvFile(""); found != "" {
		t.Errorf("FindEnvFile(\"\") should return empty string, got %v", found)
	}
	
	// Should return empty string if .env doesn't exist
	otherDir, _ := os.MkdirTemp("", "podlift-env-test-other-*")
	defer os.RemoveAll(otherDir)
	otherConfig := filepath.Join(otherDir, "podlift.yml")
	if found := FindEnvFile(otherConfig); found != "" {
		t.Errorf("FindEnvFile() should return empty when .env missing, got %v", found)
	}
}

func TestCustomEnvFile(t *testing.T) {
	// Create temp directory with custom env file
	tmpdir, err := os.MkdirTemp("", "podlift-env-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create custom env file
	customEnvPath := filepath.Join(tmpdir, "custom.env")
	os.WriteFile(customEnvPath, []byte("CUSTOM_VAR=custom_value\n"), 0644)

	// Create config
	config := &Config{
		EnvFile:    customEnvPath,
		configPath: filepath.Join(tmpdir, "podlift.yml"),
	}

	// Should load custom env file
	if err := config.SubstituteConfigEnvVars(); err != nil {
		t.Errorf("SubstituteConfigEnvVars() with custom env_file failed: %v", err)
	}

	// Check variable was loaded
	if os.Getenv("CUSTOM_VAR") != "custom_value" {
		t.Errorf("Custom env variable not loaded, got: %v", os.Getenv("CUSTOM_VAR"))
	}
	
	// Clean up
	os.Unsetenv("CUSTOM_VAR")
	
	// Test with non-existent custom file
	config.EnvFile = filepath.Join(tmpdir, "nonexistent.env")
	if err := config.SubstituteConfigEnvVars(); err == nil {
		t.Error("SubstituteConfigEnvVars() should error with non-existent env_file")
	}
	
	// Test with tilde expansion
	home, _ := os.UserHomeDir()
	config.EnvFile = "~/custom.env"
	// Create file in home directory
	homeEnvPath := filepath.Join(home, "custom.env")
	os.WriteFile(homeEnvPath, []byte("HOME_VAR=home_value\n"), 0644)
	defer os.Remove(homeEnvPath)
	
	if err := config.SubstituteConfigEnvVars(); err != nil {
		t.Errorf("SubstituteConfigEnvVars() with tilde path failed: %v", err)
	}
	os.Unsetenv("HOME_VAR")
}

