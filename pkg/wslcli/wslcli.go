package wslcli

import (
	"errors"
	"os/exec"
	"regexp"
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

// GetWSLIP returns the IP address of the running default WSL distro
func GetWSLIP() (string, error) {
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

// GetHostIP returns the IP address of Hyper-V Switch on the host connected to WSL
func GetHostIP() (string, error) {
	cmd := exec.Command("netsh", "interface", "ip", "show", "address", "vEthernet (WSL)") //, "|", "findstr", "IP Address", "|", "%", "{", "$_", "-replace", "IP Address:", "", "}", "|", "%", "{", "$_", "-replace", " ", "", "}")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ipRegex := regexp.MustCompile("IP Address:\040*(.*)\r\n")
	ipString := ipRegex.FindStringSubmatch(string(out))
	if len(ipString) != 2 {
		return "", errors.New(`netsh interface ip show address "vEthernet (WSL)"`)
	}
	return ipString[1], nil
}
