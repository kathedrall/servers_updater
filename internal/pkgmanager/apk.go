package pkgmanager

import (
	"servers_updater/internal/domain"
	"strings"
)

const (
	apkCheck     = `sudo apk list -u`
	apkGetUpdate = `sudo apk upgrade --no-cache`
	apkCount     = `sudo apk list -u | wc -l`
	apkRestart   = `if [ "$(uname -r)" != $(cat /proc/sys/kernel/osrelease 2>/dev/null) ]; then echo "YES"; else echo "NOT"; fi`
)

// ApkManager Implements interface
type ApkManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *ApkManager) GetCheckCommand() string {
	return apkCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *ApkManager) GetUpdateCommand() string {
	return apkGetUpdate
}

// GetCountPackagesCommand Return command for count packages they need updates
func (m *ApkManager) GetCountPackagesCommand() string {
	return apkCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *ApkManager) GetRestartCheckCommand() string {
	return apkRestart
}

// ParseOutput processa a saida de texto bruto do comando de simulacao
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
