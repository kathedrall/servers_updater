package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

type EmergeManager struct{}

func (m *EmergeManager) GetCheckCommand() string {
	return CommandCheckUpdate[5]
}

func (m *EmergeManager) GetInstallCommand() string {
	return CommandUpdate[5]
}

func (m *EmergeManager) ParseAptOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`\[ebuild\s+([A-Z]+)\s*\]\s+(\S+)`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			fullString := matches[2]
			lastDash := strings.LastIndex(fullString, "-")
			name := fullString
			version := "?"

			if lastDash != -1 {
				name = fullString[:lastDash]
				version = fullString[lastDash+1:]
			}

			pkgs = append(pkgs, domain.Package{
				Name:           name,
				CurrentVersion: "?",
				NewVersion:     version,
			})
		}
	}
	return pkgs, nil
}
