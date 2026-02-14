package pkgmanager

import (
	"errors"
	"fmt"
	"reflect"
	"servers_updater/internal/domain"
	"strings"
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

func TestGetManager(t *testing.T) {
	tests := []struct {
		osName        string
		expectedType  string
		expectedError bool
	}{
		{"ubuntu", "*pkgmanager.AptManager", false},
		{"debian", "*pkgmanager.AptManager", false},
		{"fedora", "*pkgmanager.DnfManager", false},
		{"alpine", "*pkgmanager.ApkManager", false},
		{"opensuse", "*pkgmanager.ZypperManager", false},
		{"arch", "*pkgmanager.PacmanManager", false},
		{"gentoo", "*pkgmanager.EmergeManager", false},
		{"freebsd", "*pkgmanager.FreebsdManager", false},
		{"windows", "", true},
	}

	for _, tt := range tests {
		t.Run("OS: "+tt.osName, func(t *testing.T) {
			manager, err := GetManager(tt.osName)
			if tt.expectedError {
				if err == nil {
					t.Errorf("There should have been an error for %s, but it came out as nil.", tt.osName)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", tt.osName, err)
			}

			managerType := reflect.TypeOf(manager).String()
			if managerType != tt.expectedType {
				t.Errorf("wrong type. Expected %s, received %s", tt.expectedType, managerType)
			}
		})
	}
}

func TestParseOutputStrategies(t *testing.T) {
	tests := []struct {
		name          string
		manager       PackageManager
		mockOutput    string
		expectedPkg   string
		expectedOld   string
		expectedNew   string
		expectedCount int
	}{
		{
			name:    "APT parser",
			manager: &AptManager{},
			mockOutput: `Inst libssl1.1 [1.1.1f-1ubuntu2] (1.1.1f-1ubuntu2.16 Ubuntu:20.04/
			focal-updates [amd64])
		    Inst curl [7.68.0-1ubuntu2.14] (7.68.0-1ubuntu2.15 Ubuntu:20.04/focal-updates [amd64])`,
			expectedPkg:   "libssl1.1",
			expectedOld:   "1.1.1f-1ubuntu2",
			expectedNew:   "1.1.1f-1ubuntu2.16",
			expectedCount: 1,
		},
		{
			name:    "DNF Parser",
			manager: &DnfManager{},
			mockOutput: `
Last metadata expiration check: 0:54:12 ago on Tue 12.
kernel.x86_64                        5.14.0-362.8.1.e9_3                        baseos
curl.x86_64                          7.76.1-26.el9_3                             appstream`,
			expectedPkg:   "kernel.x86_64",
			expectedOld:   "?",
			expectedNew:   "5.14.0-362.8.1.e19_3",
			expectedCount: 1,
		},
		{
			name:    "APK Parser",
			manager: &ApkManager{},
			mockOutput: `busybox-1.36.1-r15 x86_64 {busybox} (GPL-2.0-only) 
			ssl_client-1.36.1-r15 x86_64 {busybox} (GPL-2.0-only)`,
			expectedPkg:   "busybox-1.36.1-r15",
			expectedOld:   "?",
			expectedNew:   "latest",
			expectedCount: 1,
		},
		{
			name:    "Zypper Parser",
			manager: &ZypperManager{},
			mockOutput: `S | Repository  | Name        | Current Version | Available Version | Arch
						--+-------------+-------------+-----------------+-------------------+-------
						 v | Main Repo   | curl        | 7.60.0-1.1      | 7.61.0-2.1        | x86_64
						 v | Update Repo | libzypp     | 17.25.0-1.1     | 17.31.2-1.2       | x86_64`,

			expectedPkg:   "curl",
			expectedOld:   "7.60.0-1.1",
			expectedNew:   "7.61.0-2.1",
			expectedCount: 1,
		},
		{
			name:    "PACMAN Parser",
			manager: &PacmanManager{},
			mockOutput: `coreutils 8.32-1 -> 9.0-1
libsystemd 249.4-1 -> 249.5-1`,
			expectedPkg:   "coreutils",
			expectedOld:   "8.32.1",
			expectedNew:   "9.0-1",
			expectedCount: 1,
		},
		{
			name:    "EMERGE Parser",
			manager: &EmergeManager{},
			mockOutput: `[ebuild     U ] net-misc/curl-7.79.1 [7.78.0] USE="ssl -ldap"
						 [ebuild  N    ] app-editors/vim-8.2.3456`,
			expectedPkg:   "net-misc/curl",
			expectedOld:   "?",
			expectedNew:   "7.79.1",
			expectedCount: 1,
		},
		{
			name:    "FREEBSD Parser",
			manager: &FreebsdManager{},
			mockOutput: `Updating FreeBSD repository catalogue...
The following 2 package(s) will be affected (of 0 checked):
Installed packages to be UPGRADED:
	curl: 8.4.0 -> 8.5.0
	bash: 5.2.15 -> 5.2.21`,
			expectedPkg:   "curl",
			expectedOld:   "8.4.0",
			expectedNew:   "8.5.0",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs, err := tt.manager.ParseOutput(tt.mockOutput)
			//t.Logf("pkgs %s", pkgs)

			if err != nil {
				t.Fatalf("Parser err: %v", err)
			}

			if len(pkgs) != tt.expectedCount {
				t.Errorf("Incorrect Count. Expected %d, find out %d", tt.expectedCount, len(pkgs))
				return
			}
			if len(pkgs) > 0 {
				p := pkgs[0]
				if !strings.Contains(p.Name, tt.expectedPkg) {
					t.Errorf("Wrong package error. I was waiting for %s, but returned '%s'", tt.expectedPkg, p.Name)
				}
				if p.NewVersion != tt.expectedNew {
					t.Errorf("New incorrect version. I expected '%s' but returned '%s'", tt.expectedNew, p.NewVersion)
				}
				if tt.expectedOld != "?" && p.CurrentVersion != tt.expectedOld {
					t.Errorf("Old incorrect version. I expected '%s' but returned '%s", tt.expectedOld, p.CurrentVersion)
				}
			}
		})
	}
}

func TestDetectOs(t *testing.T) {
	client := []struct {
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

	for _, tt := range client {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSSHClient{Output: tt.mockOutput}
			machine := &domain.Machine{}

			err := DetectOs(mock, machine)
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

	t.Run("Error Detect OS Name", func(t *testing.T) {
		mock := &MockSSHClient{Err: errors.New("connection failed")}
		machine := &domain.Machine{}
		err := DetectOs(mock, machine)
		if err != nil {
			t.Errorf("Expected error, got nil")
		}
	})

}

func TestCommandsStanityCheck(t *testing.T) {
	managers := []PackageManager{
		&AptManager{},
		&DnfManager{},
		&ApkManager{},
		&ZypperManager{},
		&PacmanManager{},
		&EmergeManager{},
		&FreebsdManager{},
	}

	for _, m := range managers {
		cmdCheck := m.GetCheckCommand()
		cmdInstall := m.GetUpdateCommand()

		if cmdCheck == "" {
			t.Errorf("Empty check command for %T", m)
		}
		if cmdInstall == "" {
			t.Errorf("Empty check coommand for %T", m)
		}

		if !strings.HasPrefix(cmdCheck, "sudo") {
			t.Logf("Warning: The command %T does begin with sudo: %s:", m, cmdCheck)
		}

	}
}
