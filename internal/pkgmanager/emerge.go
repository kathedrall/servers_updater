package pkgmanager

import (
	"regexp"
	"servers_updater/internal/domain"
	"strings"
)

const (
	emergeCheck     = `sudo emerge -pvu @world`
	emergeGetUpdate = `sudo emerge -u --auto-umask-write @world`
	emergeCount     = `sudo emerge -pvuND @world | grep -c "ebuild"`
	emergeRestart   = `if [ "$(uname -r)" != "$(eselect kernel show | grep -o 'linux-[0-9].*')" ]; then echo "YES"; else echo "NOT"; fi`
)

// EmergeManager Implements interface
type EmergeManager struct{}

// GetCheckCommand Return command for updater simulator and packages list
func (m *EmergeManager) GetCheckCommand() string {
	return emergeCheck
}

// GetUpdateCommand Return command for application silent updates
func (m *EmergeManager) GetUpdateCommand() string {
	return emergeGetUpdate
}

// GetCountPackagesCommand Return command for application silent updates
func (m *EmergeManager) GetCountPackagesCommand() string {
	return emergeCount
}

// GetRestartCheckCommand Check if a system restart is required
func (m *EmergeManager) GetRestartCheckCommand() string {
	return dnfRestart
}

// ParseOutput Process text output for comannd simulation
func (m *EmergeManager) ParseOutput(output string) ([]domain.Package, error) {
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
