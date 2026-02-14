package ssh

import (
	"fmt"
	"os"
	"path/filepath"
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

	configContent := `Host alias-test
		HostName 192.168.0.20
		User user-test
		Port 2222
		IdentityFile ˜/.ssh/key_test.pem`

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write mock config: %v", err)
	}
	t.Setenv("HOME", tmpHome)

	machine := &domain.Machine{
		Host: "alias-test",
	}

	// Call resolveHostConfig to resolve the SSH config
	resolveHostConfig(machine)

	if machine.Host != "192.168.0.20" {
		t.Errorf("Hostname resolution failed. Expected 192.168.0.20, got %s", machine.Host)
	}

	if machine.User != "" && machine.User != "user-test" {
		t.Errorf("User resolution failed. Expected user-test got %v", machine.User)
	}

	if machine.Port != 0 && machine.Port != 2222 {
		t.Errorf("Port resolution failed. Expected 2222, got %v", machine.Port)
	}

	expectedKeyPath := filepath.Join(tmpHome, ".ssh", "key_test.pem")
	if machine.KeyPath != "" && machine.KeyPath != expectedKeyPath {
		t.Errorf("IdentityFile resolution failed. \nExpected: %s\nGot: %v", expectedKeyPath, machine.KeyPath)
	}
}
