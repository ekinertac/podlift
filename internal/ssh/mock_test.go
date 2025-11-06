package ssh

import (
	"bytes"
	"strings"
	"testing"
)

func TestMockClient_Execute(t *testing.T) {
	mock := NewMockClient()
	
	// Test default docker version response
	output, err := mock.Execute("docker --version")
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if !strings.Contains(output, "Docker version") {
		t.Errorf("Execute() should return Docker version, got: %v", output)
	}
}

func TestMockClient_CustomExecute(t *testing.T) {
	mock := &MockClient{
		ExecuteFunc: func(cmd string) (string, error) {
			return "custom response", nil
		},
	}

	output, err := mock.Execute("test")
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if output != "custom response" {
		t.Errorf("Execute() = %v, want 'custom response'", output)
	}
}

func TestMockClient_ExecuteWithOutput(t *testing.T) {
	mock := NewMockClient()
	
	var stdout, stderr bytes.Buffer
	err := mock.ExecuteWithOutput("test command", &stdout, &stderr)
	if err != nil {
		t.Errorf("ExecuteWithOutput() error = %v", err)
	}
}

func TestMockClient_TestConnection(t *testing.T) {
	mock := NewMockClient()
	
	err := mock.TestConnection()
	if err != nil {
		t.Errorf("TestConnection() error = %v", err)
	}
}

func TestMockClient_CheckDocker(t *testing.T) {
	mock := NewMockClient()
	
	version, err := mock.CheckDocker()
	if err != nil {
		t.Errorf("CheckDocker() error = %v", err)
	}
	if version == "" {
		t.Error("CheckDocker() should return version")
	}
}

func TestMockClient_CheckPort(t *testing.T) {
	mock := NewMockClient()
	
	available, err := mock.CheckPort(8000)
	if err != nil {
		t.Errorf("CheckPort() error = %v", err)
	}
	if !available {
		t.Error("CheckPort() should return true by default")
	}
}

func TestMockClient_GetDiskSpace(t *testing.T) {
	mock := NewMockClient()
	
	space, err := mock.GetDiskSpace()
	if err != nil {
		t.Errorf("GetDiskSpace() error = %v", err)
	}
	if space <= 0 {
		t.Errorf("GetDiskSpace() = %v, want > 0", space)
	}
}

func TestMockClient_CheckExistingService(t *testing.T) {
	mock := NewMockClient()
	
	exists, info, err := mock.CheckExistingService("test")
	if err != nil {
		t.Errorf("CheckExistingService() error = %v", err)
	}
	if exists {
		t.Error("CheckExistingService() should return false by default")
	}
	if info != nil {
		t.Error("CheckExistingService() should return nil info by default")
	}
}

func TestMockClient_CopyFile(t *testing.T) {
	mock := NewMockClient()
	
	err := mock.CopyFile("/local/file", "/remote/file")
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
	}
}

func TestMockClient_CopyFileWithProgress(t *testing.T) {
	mock := NewMockClient()
	progressCalled := false
	
	err := mock.CopyFileWithProgress("/local/file", "/remote/file", func(sent, total int64) {
		progressCalled = true
		if sent != 100 || total != 100 {
			t.Errorf("Progress = %d/%d, want 100/100", sent, total)
		}
	})
	
	if err != nil {
		t.Errorf("CopyFileWithProgress() error = %v", err)
	}
	if !progressCalled {
		t.Error("Progress callback should be called")
	}
}

func TestMockClient_Connect(t *testing.T) {
	mock := &MockClient{}
	
	err := mock.Connect()
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}
	if !mock.Connected {
		t.Error("Connected should be true after Connect()")
	}
}

func TestMockClient_CustomFunction(t *testing.T) {
	mock := &MockClient{
		CheckDockerFunc: func() (string, error) {
			return "Custom version", nil
		},
	}

	version, err := mock.CheckDocker()
	if err != nil {
		t.Errorf("CheckDocker() error = %v", err)
	}
	if version != "Custom version" {
		t.Errorf("CheckDocker() = %v, want 'Custom version'", version)
	}
}

