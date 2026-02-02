package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

type ZypperManager struct{}

func (m *ZypperManager) GetCheckCommand() string {
	return CommandCheckUpdate[3]
}

func (m *ZypperManager) GetInstallCommand() string {
	return CommandUpdate[3]
}

func (m *ZypperManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`\|\s+([^\s|]+)\s+\|\s+([^\s|]+)\s+\|\s+([^\s|]+)\s+\|\s+[^\s|]+\s*$`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "repository") || strings.HasPrefix(line, "--") {
			continue
		}

		matches := re.FindStringSubmatch(line)
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
