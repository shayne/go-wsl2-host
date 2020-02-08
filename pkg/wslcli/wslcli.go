package wslcli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// RunningDistros returns list of distros names running
func RunningDistros() ([]string, error) {
	cmd := exec.Command("wsl.exe", "-l", "-q", "--running")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	decoded, err := decodeOutput(out)
	if err != nil {
		return nil, errors.New("failed to decode output")
	}
	return strings.Split(decoded, "\r\n"), nil
}

// ListAll returns output for "wsl.exe -l -v"
func ListAll() (string, error) {
	cmd := exec.Command("wsl.exe", "-l", "-v")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wsl -l -v failed: %w", err)
	}
	decoded, err := decodeOutput(out)
	if err != nil {
		return "", fmt.Errorf("failed to decode output: %w", err)
	}
	return decoded, nil
}

// GetIP returns the IP address of the given distro
// Suggest check if running before calling this function as
// it has the side-effect of starting the distro
func GetIP(name string) (string, error) {
	// check if ~/.wsl2hosts is exist
	cmd := exec.Command("wsl.exe", "-d", name, "--", "eval", "ls -a ~ | grep .wsl2hosts")
	out, err := cmd.Output()
	sout := string(out)
	sout = strings.TrimSpace(sout)
	if err == nil && sout == ".wsl2hosts" {
		// run ~/.wsl2hosts as a bash script
		// output data format:
		// first line is VM's ip: xxx.xxx.xxx.xxx
		// second line is host alias or empty: arch.wsl
		cmd := exec.Command("wsl.exe", "-d", name, "--", "sh", "~/.wsl2hosts")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("RunCommand failed: %w", err)
		}
		sout := string(out)
		sout = strings.TrimSpace(sout)
		lines := strings.Split(sout, "\n")
		if sout == "" || len(lines) == 0 {
			return "", errors.New("invalid output from .wsl2hosts")
		}
		line := lines[0]
		line = strings.TrimSpace(line)
		ips := strings.Split(line, " ")
		if line == "" || len(ips) == 0 {
			return "", errors.New("invalid output from .wsl2hosts")
		}
		return ips[0], nil
	} else {
		cmd := exec.Command("wsl.exe", "-d", name, "--", "hostname", "-I")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		sout := string(out)
		sout = strings.TrimSpace(sout)
		ips := strings.Split(sout, " ")
		if sout == "" || len(ips) == 0 {
			return "", errors.New("invalid output from hostname -I")
		}
		// first IP is the correct interface
		return ips[0], nil
	}
}

// RunCommand runs the given command via `bash -c` under
// the default WSL distro
func RunCommand(command string, args ...string) (string, error) {
	cmdstr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	cmd := exec.Command("wsl.exe", "--", "bash", "-c", cmdstr)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	sout := string(out)
	return sout, nil
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

func decodeOutput(raw []byte) (string, error) {
	win16le := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	utf16bom := unicode.BOMOverride(win16le.NewDecoder())
	unicodeReader := transform.NewReader(bytes.NewReader(raw), utf16bom)
	decoded, err := ioutil.ReadAll(unicodeReader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
