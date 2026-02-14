package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
)

const (
	aptCheck     = `sudo apt-get upgrade -s`
	aptGetUpdate = `sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -y`
	aptCount     = `sudo apt-get -s -o Debug::NoLocking=true upgrade | grep -c ^Inst`
	aptRestart   = `[ -f /var/run/reboot-required ] && echo '"SIM"' || '"NAO"`
)

// AptManager Implements interface
type AptManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *AptManager) GetCheckCommand() string {
	return aptCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *AptManager) GetUpdateCommand() string {
	return aptGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *AptManager) GetCountPackagesCommand() string {
	return aptCount
}

// GetRestartCheckCommand check if a system restart is required
func (m *AptManager) GetRestartCheckCommand() string {
	return aptRestart
}

// ParseOutput Process text output for comannd simulation
func (m *AptManager) ParseOutput(output string) ([]domain.Package, error) {
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
