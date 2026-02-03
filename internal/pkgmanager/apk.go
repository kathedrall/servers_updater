package pkgmanager

import (
	"servers_updater/internal/domain"
	"strings"
)

type ApkManager struct{}

func (m *ApkManager) GetCheckCommand() string {
	return CommandCheckUpdate[1]
}

func (m *ApkManager) GetInstallCommand() string {
	return CommandUpdate[1]
}

func (m *ApkManager) ParseOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}

		fulName := parts[0]
		pkgs = append(pkgs, domain.Package{
			Name:           fulName,
			NewVersion:     "latest",
			CurrentVersion: "?",
		})
	}
	return pkgs, nil
}
