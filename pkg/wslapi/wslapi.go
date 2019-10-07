package wslapi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/shayne/go-wsl2-host/pkg/wslcli"
)

// DistroInfo data structure for state of a WSL distro
type DistroInfo struct {
	Name    string
	Running bool
	Version int
	Default bool
}

// GetAllInfo checks all distros and returns slice
// of state info for all
func GetAllInfo() ([]*DistroInfo, error) {
	output, err := wslcli.ListAll()
	if err != nil {
		return nil, fmt.Errorf("wsl list all failed: %w", err)
	}

	lines := strings.Split(output, "\r\n")

	if len(lines) <= 1 {
		return nil, errors.New("bad output from wslcli, cannot parse")
	}

	lines = lines[1:] // skip first header line

	var infos []*DistroInfo
	for _, line := range lines {
		info := &DistroInfo{}
		if len(line) <= 0 {
			continue
		}
		if line[0] == '*' {
			line = line[1:]
			info.Default = true
		}
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid field length for distro: %q", line)
		}
		info.Name = fields[0]
		info.Running = fields[1] == "Running"
		version, err := strconv.ParseInt(fields[2], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid version for distro: %q", line)
		}
		info.Version = int(version)

		infos = append(infos, info)
	}

	return infos, nil
}

// IsRunning returns whether or not a given WSL distro is running
func IsRunning(name string) (bool, error) {
	running, err := wslcli.RunningDistros()
	if err != nil {
		return false, fmt.Errorf("running distros failed: %w", err)
	}
	for _, distro := range running {
		if distro == name {
			return true, nil
		}
	}
	return false, nil
}

// GetIP returns the IP address of a running WSL distro
func GetIP(name string) (string, error) {
	running, err := IsRunning(name)
	if err != nil {
		return "", fmt.Errorf("IsRunning failed: %w", err)
	}
	if running {
		return wslcli.GetIP(name)
	}
	return "", fmt.Errorf("GetIP failed, distro '%s' is not running", name)
}
