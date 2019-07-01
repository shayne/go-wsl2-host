package wslcli

import (
	"errors"
	"os/exec"
	"strings"
)

// Running returns bool, error whether or not WSL instance is running
func Running() (bool, error) {
	cmd := exec.Command("wsl.exe", "-l", "-q", "--running")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(out) != 0, nil
}

// IP returns the IP address of the running default WSL distro
func IP() (string, error) {
	cmd := exec.Command("wsl.exe", "--", "hostname", "-I")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	sout := string(out)
	ips := strings.Split(sout, " ")
	if len(ips) == 0 {
		return "", errors.New("invalid output from hostname -I")
	}
	return ips[0], nil
}
