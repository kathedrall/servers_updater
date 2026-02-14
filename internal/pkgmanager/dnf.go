package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

const (
	dnfCheck     = `sudo dnf check-update`
	dnfGetUpdate = `sudo dnf update -y`
	dnfCount     = `sudo dnf check-update --quiet | grep -v '^$' | wc -l`
	dnfRestart   = `needs-restarting -r > /dev/null 2>&1; if [ $? -eq 1]; then echo "YES"; else echo "NOT"; fi`
)

// DnfManager Implements interface
type DnfManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *DnfManager) GetCheckCommand() string {
	return dnfCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *DnfManager) GetUpdateCommand() string {
	return dnfGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *DnfManager) GetCountPackagesCommand() string {
	return dnfCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *DnfManager) GetRestartCheckCommand() string {
	return dnfRestart
}

// ParseOutput Process text output for comannd simulation
func (m *DnfManager) ParseOutput(output string) ([]domain.Package, error) {
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
