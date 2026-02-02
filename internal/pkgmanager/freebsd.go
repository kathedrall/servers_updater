package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

type FreebsdManager struct{}

func (m *FreebsdManager) GetCheckCommand() string {
	return CommandCheckUpdate[6]
}

func (m *FreebsdManager) GetInstallCommand() string {
	return CommandUpdate[6]
}

func (m *FreebsdManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`\s*([ˆ\s:]+):\s+([ˆ\s]+)\s+->\s+([ˆ\s]+)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "The following") || strings.Contains(line, "Numver of packages") {
			continue
		}

		cleanLine := strings.TrimPrefix(strings.TrimSpace(line), "Upgrading")
		matches := re.FindStringSubmatch(cleanLine)
		if len(matches) >= 4 {
			pkgs = append(pkgs, domain.Package{
				Name:           matches[1],
				CurrentVersion: matches[2],
				NewVersion:     matches[3],
			})

		}
	}
	return pkgs, nil
}
