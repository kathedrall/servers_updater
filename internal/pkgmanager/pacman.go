package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

const (
	pacmanCheck     = `sudo pacman -Sy && pacman -Qu`
	pacmanGetUpdate = `sudo pacman -Syu --noconfirm`
	pacmanCount     = `sudo checkupdates | wc -l`
	pacmanRestart   = `running=$(uname -r); installed=$(pacman -Q linux | awk '{print $2}'); if [[ "$running" != *"$installed"* ]]; then echo "YES"; else echo "NOT"; fi`
)

// PacmanManager Implements interface
type PacmanManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *PacmanManager) GetCheckCommand() string {
	return pacmanCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *PacmanManager) GetUpdateCommand() string {
	return pacmanGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *PacmanManager) GetCountPackagesCommand() string {
	return pacmanCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *PacmanManager) GetRestartCheckCommand() string {
	return pacmanRestart
}

// ParseOutput Process text output for comannd simulation
func (m *PacmanManager) ParseOutput(output string) ([]domain.Package, error) {
	var pkgs []domain.Package

	re := regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s+->\s+(\S+)`)
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
