package wslapi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"regexp"

	"github.com/shayne/go-wsl2-host/pkg/wslcli"
)

const dockerDesktopDistros = "docker-desktop"

// DistroInfo data structure for state of a WSL distro
type DistroInfo struct {
	Name    string
	Running bool
	Version int
	Default bool
	IP      string
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
		if strings.HasPrefix(info.Name, dockerDesktopDistros) {
			continue
		}
		info.Running = fields[1] == "Running"
		version, err := strconv.ParseInt(fields[2], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid version for distro: %q", line)
		}
		info.Version = int(version)

		if info.Version == 1 {
			info.IP = "127.0.0.1"
		} else if info.Running {
			info.IP, err = GetIP(info.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get IP for distro %q: %v", info.Name, err)
			}
		}

		infos = append(infos, info)
	}

	return infos, nil
}

func Shutdown() error {
	return wslcli.Shutdown()
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
	out, err := wslcli.RunCommand("cat", "~/.wsl2hosts")
	if err != nil {
		return nil, fmt.Errorf("RunCommand failed: %w", err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, fmt.Errorf("no host aliases")
	}
	return regexp.MustCompile("\\s+").Split(out, -1), nil
}

func GetHostIP(distro string, host string) (string, error) {
	return wslcli.GetHostIPFromHosts(distro, host)
}

func UpdateHostIP(distro string, host string, ip string) error {
	return wslcli.UpdateHostIP(distro, host, ip)
}

func AddHostIP(distro string, host string, ip string) error {
	return wslcli.AddHostIP(distro, host, ip)
}

func DeleteHost(distro string, host string) error {
	return wslcli.DeleteHost(distro, host)
}

func AddOrUpdateHostIP(distro string, host string, ip string) error {
	old_ip, err := GetHostIP(distro, host)
	if err != nil {
		return err
	}
	if old_ip == ip {
		return nil
	}
	if len(old_ip) == 0 {
		return AddHostIP(distro, host, ip)
	} else if old_ip != ip {
		return UpdateHostIP(distro, host, ip)
	}
	return nil
}
