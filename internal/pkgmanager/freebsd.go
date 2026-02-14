package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

const (
	pkgCheck     = `sudo pkg upgrade -n`
	pkgGetUpdate = `sudo pkg upgrade -y`
	pkgCount     = `pkg version -l "<" | wc -l`
	pkgRestart   = `if [ "$(freebsd-version -k)" != "$(uname -r)" ]; then echo "YES"; else echo "NOT"`
)

// FreebsdManager Implements interface
type FreebsdManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *FreebsdManager) GetCheckCommand() string {
	return pkgCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *FreebsdManager) GetUpdateCommand() string {
	return pkgGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *FreebsdManager) GetCountPackagesCommand() string {
	return pkgCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *FreebsdManager) GetRestartCheckCommand() string {
	return pkgRestart
}

// ParseOutput Process text output for comannd simulation
func (m *FreebsdManager) ParseOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`^\s*(\S+):\s+(\S+)\s+->\s+(\S+)`)
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
