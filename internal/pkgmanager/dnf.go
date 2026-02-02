package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

type DnfManager struct{}

func (m *DnfManager) GetCheckCommand() string {
	return CommandCheckUpdate[0]
}

func (m *DnfManager) GetInstallCommand() string {
	return CommandUpdate[0]
}

func (m *DnfManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Last metadata expiration check") || line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			pkgs = append(pkgs, domain.Package{
				Name:           matches[1],
				CurrentVersion: "?",
				NewVersion:     matches[2],
			})
		}
	}
	return pkgs, nil
}
