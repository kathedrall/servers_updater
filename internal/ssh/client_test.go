package ssh

import (
	"os"
	"path/filepath"
	"errors"
	"fmt"
	"servers_updater/internal/domain"
	"testing"
)

type MockSSHClient struct {
	Output string
	Err    error
}

func (m *MockSSHClient) ExecuteCommand(cmd string) (string, error) {
	fmt.Printf("[DEBUG MOCK] command received: %s | return: %s (Err: %v)\n", cmd, m.Output, m.Err)
	return m.Output, m.Err
}

func (m *MockSSHClient) Close() error {
	return nil
}

func TestResolveHostConfig(t *testing.T) {
	tmpHome := t.TempDir()
	sshDir := filepath.Join(tmpHome, ".ssh")

	if err := os.Mkdir(sshDir, 0700); err != nil {
		t.Fatalf("Failed to create tem .ssh dir: %sv", err)
	}

	configContent :=  `Host alias-test
		HostName 192.168.0.20
		User user-test
		Port 2222
		IdentityFile ˜/.ssh/key_test.pem`

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write mock config: %v", err)
	}		
	t.Setenv("HOME", tmpHome)

	machine := &domain.Machine {
		Host: "alias-test",
		HostName: "user-test",
		Port: 2222,
	}
	if machine.Host != "192.168.0.20" {
		t.Errorf("Hostname resolution failed. Expected 192.168.0.10, got %s", machine.Host)
	}

	if machine.User != "user-test" {
		t.Errorf("User resolution failed. Expected user-test got %s", machine.User)
	}

	if machine.Port != 2222 {
		t.Errorf("Port resolution failed. Expected 2222, got %d", machine.Port)
	}

	expectedKeyPath := filepath.Join(tmpHome, ".ssh", "key_test.pem")
	if machine.KeyPath != expectedKeyPath {
		t.Errorf("IdentityFile resolution failed. \nExpected: %s\nGot: %s", expectedKeyPath, machine.KeyPath)
	}	
}

func TestIdentifyOs(t *testing.T) {
	tests := []struct {
		name            string
		mockOutput      string
		expectedOS      string
		expectedPretty  string
		expectedSupport bool
	}{
		{
			name:            "Detect Ubuntu",
			mockOutput:      "PRETTY_NAME=\"Ubuntu 22.04 LTS\"\nID=ubuntu",
			expectedOS:      "ubuntu",
			expectedPretty:  "Ubuntu 22.04 LTS",
			expectedSupport: true,
		},
		{
			name:            "Detect  Unsupported Alpine",
			mockOutput:      "PRETTY_NAME=\"Alpine Linux\"\nID=alpine",
			expectedOS:      "alpine",
			expectedPretty:  "Alpine Linux",
			expectedSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSSHClient{Output: tt.mockOutput}
			machine := &domain.Machine{}

			err := IdentifyOS(mock, machine)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if machine.OSName != tt.expectedOS {
				t.Errorf("OS: got %s, want %s", machine.OSName, tt.expectedOS)
			}

			if machine.IsSupported != tt.expectedSupport {
				t.Errorf("Support: got %v, want %v", machine.IsSupported, tt.expectedSupport)
			}
		})
	}
	t.Run("Handle SSH Error", func(t *testing.T) {
		mock := &MockSSHClient{Err: errors.New("connection failed")}
		machine := &domain.Machine{}
		err := IdentifyOS(mock, machine)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
