package wslapi

import (
	"errors"
	"fmt"
	"os/exec"
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

// GetDefaultDistro returns the info for the default distro
func GetDefaultDistro() (*DistroInfo, error) {
	infos, err := GetAllInfo()
	if err != nil {
		return nil, fmt.Errorf("GetAllInfo failed: %w", err)
	}
	for _, i := range infos {
		if i.Default == true {
			return i, nil
		}
	}
	return nil, errors.New("failed to find default")
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
	if !running {
		return "", fmt.Errorf("GetIP failed, distro '%s' is not running", name)
	}
	return wslcli.GetIP(name)
}

// GetHostAliases returns custom hosts referenced in `~/.wsl2hosts`
// of default WSL distro
func GetHostAliases() ([]string, error) {
	info, err := GetDefaultDistro()
	if err != nil {
		return nil, fmt.Errorf("GetDefaultDistro failed: %w", err)
	}
	if !info.Running {
		return nil, errors.New("default distro not running")
	}
	// check if ~/.wsl2hosts is exist
	cmd := exec.Command("wsl.exe", "--", "eval", "ls -a ~ | grep .wsl2hosts")
	out, err := cmd.Output()
	sout := string(out)
	sout = strings.TrimSpace(sout)
	if err == nil && sout == ".wsl2hosts" {
		// run ~/.wsl2hosts as a bash script
		// output data format:
		// first line is VM's ip: xxx.xxx.xxx.xxx
		// second line is host alias or empty: arch.wsl
		cmd := exec.Command("wsl.exe", "--", "sh", "~/.wsl2hosts")
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("RunCommand failed: %w", err)
		}
		sout := string(out)
		sout = strings.TrimSpace(sout)
		lines := strings.Split(sout, "\n")
		if sout == "" || len(lines) == 0 {
			return nil, errors.New("invalid output from .wsl2hosts")
		}
		if len(lines) == 1 {
			return nil, fmt.Errorf("no host aliases")
		}
		line := lines[1]
		line = strings.TrimSpace(line)
		return strings.Split(line, " "), nil
	} else if err != nil {
		return nil, fmt.Errorf("RunCommand failed: %w", err)
	}
	return nil, fmt.Errorf("no host aliases")
}
