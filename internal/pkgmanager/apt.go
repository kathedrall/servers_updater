package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
)

type AptManager struct{}

func (m *AptManager) GetCheckCommand() string {
	return CommandCheckUpdate[2]
}

func (m *AptManager) GetInstallCommand() string {
	return CommandUpdate[2]
}

func (m *AptManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package
	re := regexp.MustCompile(`Inst\s+([^\s]+)\s+\[([^\]]+)\]\s+\(([^\s]+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, m := range matches {
		pkgs = append(pkgs, domain.Package{
			Name:           m[1],
			CurrentVersion: m[2],
			NewVersion:     m[3],
		})
	}
	return pkgs, nil
}
