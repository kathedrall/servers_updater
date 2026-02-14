package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

const (
	zypperCheck     = `sudo zypper list-updates`
	zypperGetUpdate = `sudo zypper update --non-interactive`
	zypperCount     = `sudo zypper -q lu --best-effort | grep -c 'v |'`
	zypperRestart   = `sudo zypper ps -s >/dev/null 2>&1 && echo "NOT" || echo "YES"`
)

// ZypperManager Implements interface
type ZypperManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *ZypperManager) GetCheckCommand() string {
	return zypperCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *ZypperManager) GetUpdateCommand() string {
	return zypperGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *ZypperManager) GetCountPackagesCommand() string {
	return zypperCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *ZypperManager) GetRestartCheckCommand() string {
	return zypperRestart
}

// ParseOutput Process text output for comannd simulation
func (m *ZypperManager) ParseOutput(output string) ([]domain.Package, error) {
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
