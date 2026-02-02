package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

type PacmanManager struct{}

func (m *PacmanManager) GetCheckCommand() string {
	return CommandCheckUpdate[4]
}

func (m *PacmanManager) GetInstallCommand() string {
	return CommandUpdate[4]
}

func (m *PacmanManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`ˆ(\S+)\s+(\S+)\s+->\s+(\S+)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
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
