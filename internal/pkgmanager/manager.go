package pkgmanager

import (
	"fmt"
	"servers_updater/internal/domain"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const releaseName = "cat /etc/os-release || uname -a"

// PackageManager Implementa interface
type PackageManager interface {
	GetCheckCommand() string
	GetUpdateCommand() string
	GetCountPackagesCommand() string
	GetRestartCheckCommand() string
	ParseOutput(output string) ([]domain.Package, error)
}

// DetectOs Conecta na maquina e descobre o sistema operacional
func DetectOs(client domain.SSHClient, m *domain.Machine) error {
	output, err := client.ExecuteCommand(releaseName)
	if err != nil {
		e := fmt.Errorf("failed to detect OS: %v", err)
		return e
	}
	m.PrettyName = ""
	m.OSName = ""
	m.IsSupported = false

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\"", "")

		if strings.HasPrefix(line, "PRETTY_NAME=") {
			m.PrettyName = strings.TrimPrefix(line, "PRETTY_NAME=")
		}
		if strings.HasPrefix(line, "ID=") {
			m.OSName = strings.TrimPrefix(line, "ID=")
		}
	}

	if m.OSName == "" {
		if strings.Contains(strings.ToLower(output), "freebsd") {
			m.PrettyName = "freebsd"
			m.PrettyName = "FreeBSD"
		} else {
			m.OSName = "unknown"
		}
	}

	if m.PrettyName == "" {
		m.PrettyName = cases.Title(language.Und).String(m.OSName)
	}

	return nil
}

func GetManager(osName string) (PackageManager, error) {
	switch osName {
	case "ubuntu", "debian", "kali", "pop", "linuxmint":
		return &AptManager{}, nil
	case "fedora", "centos", "rhel", "rocky", "almalinux":
		return &DnfManager{}, nil
	case "alpine":
		return &ApkManager{}, nil
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
		return &ZypperManager{}, nil
	case "arch", "manjaro":
		return &PacmanManager{}, nil
	case "gentoo":
		return &EmergeManager{}, nil
	case "freebsd":
		return &FreebsdManager{}, nil
	default:
		e := fmt.Errorf("OS not supported: %s", osName)
		return nil, e
	}
}
