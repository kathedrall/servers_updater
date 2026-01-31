package ssh

import (
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

func TestParseAptOutput(t *testing.T) {
	mockOutput := `Inst libssl1.1 [1.1.1f-1ubuntu2] (1.1.1f-1ubuntu2.16 Ubuntu:20.04/focal-updates [amd64])
Inst nginx [1.18.0-0ubuntu1] (1.18.0-0ubuntu.1.4)
Conf nginx (1.18.0-0ubuntu1.4 Ubuntu1.4 Ubuntu:20.04/focal-updates [amd64])`

	t.Run("Validate packet extraction via regex.", func(t *testing.T) {
		pkgs := ParseAptOutput(mockOutput)
		if len(pkgs) != 2 {
			t.Errorf("I was expecting 2 packages, I found %d", len(pkgs))
		}
		if pkgs[0].Name != "libssl1.1" || pkgs[0].NewVersion != "1.1.1f-1ubuntu2.16" {
			t.Errorf("Error generating libssl parse. Name: %s, version: %s", pkgs[1].Name, pkgs[1].CurrentVersion)
		}
	})

	t.Run("No packages", func(t *testing.T) {
		pkgs := ParseAptOutput("0 upgrades, 0 newly instaled, 0 to remove")
		if len(pkgs) != 0 {
			t.Errorf("It should return an empty list.")
		}
	})
}
