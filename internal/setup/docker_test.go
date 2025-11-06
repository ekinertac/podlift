package setup

import (
	"testing"
)

func TestStdoutWriter_Write(t *testing.T) {
	w := &stdoutWriter{}
	
	tests := []struct {
		input string
		want  int
	}{
		{"test output", 11},
		{"", 0},
		{"   trimmed   ", 13},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			n, err := w.Write([]byte(tt.input))
			if err != nil {
				t.Errorf("Write() error = %v", err)
			}
			if n != tt.want {
				t.Errorf("Write() n = %d, want %d", n, tt.want)
			}
		})
	}
}

func TestStderrWriter_Write(t *testing.T) {
	w := &stderrWriter{}
	
	n, err := w.Write([]byte("error message"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 13 {
		t.Errorf("Write() n = %d, want 13", n)
	}
}

// Note: InstallDocker, ConfigureFirewall, ApplySecurity all require real SSH
// These are tested in E2E tests with Multipass

