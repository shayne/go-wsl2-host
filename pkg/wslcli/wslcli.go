package wslcli

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/bits"
	"os/exec"
	"regexp"
	"strconv"
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

func netmaskToBits(mask uint32) int {
	return bits.OnesCount32(mask)
}
func hexToUint32LE(hex string) (uint32, error) {
	i, err := strconv.ParseInt(hex[6:8]+hex[4:6]+hex[2:4]+hex[0:2], 16, 64)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

type routeInfo struct {
	net  uint32
	mask uint32
}

func getRouteInfo(name string) (*routeInfo, error) {
	cmd := exec.Command("wsl.exe", "-d", name, "--", "cat", "/proc/net/route")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	ri := &routeInfo{}
	sout := string(out)
	sout = strings.TrimSpace(sout)
	lines := strings.Split(sout, "\n")
	lines = lines[1:]
	for _, line := range lines {
		fs := strings.Fields(line)
		if ri.mask > 0 && ri.net > 0 {
			break
		}
		if fs[0] != "eth0" {
			continue
		}
		if fs[1] != "00000000" {
			net, err := hexToUint32LE(fs[1])
			if err != nil {
				return nil, fmt.Errorf("failed to convert network to Uint32: %w", err)
			}
			ri.net = net
		}
		if fs[7] != "00000000" {
			mask, err := hexToUint32LE(fs[7])
			if err != nil {
				return nil, fmt.Errorf("failed to convert netmask to Uint32: %w", err)
			}
			ri.mask = mask
		}
	}

	return ri, nil
}

func isIPInRange(ri *routeInfo, ip uint32) bool {
	return (ri.net & ri.mask) == (ip & ri.mask)
}

func ipToUint32(ip string) (uint32, error) {
	octets := strings.Split(ip, ".")
	if len(octets) != 4 {
		return 0, errors.New("invalid IP address")
	}

	var io uint32

	o1, err := strconv.Atoi(octets[0])
	if err != nil {
		return 0, fmt.Errorf("failed to parse IP address, %s: %w", ip, err)
	}
	io += uint32(o1 << 24)
	o2, err := strconv.Atoi(octets[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse IP address, %s: %w", ip, err)
	}
	io += uint32(o2 << 16)
	o3, err := strconv.Atoi(octets[2])
	if err != nil {
		return 0, fmt.Errorf("failed to parse IP address, %s: %w", ip, err)
	}
	io += uint32(o3 << 8)
	o4, err := strconv.Atoi(octets[3])
	if err != nil {
		return 0, fmt.Errorf("failed to parse IP address, %s: %w", ip, err)
	}
	io += uint32(o4)

	return io, nil
}

// GetIP returns the IP address of the given distro
// Suggest check if running before calling this function as
// it has the side-effect of starting the distro
func GetIP(name string) (string, error) {
	ri, err := getRouteInfo(name)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("wsl.exe", "-d", name, "--", "cat", "/proc/net/fib_trie")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	sout := string(out)
	sout = strings.TrimSpace(sout)
	if sout == "" {
		return "", errors.New("invalid output from fib_trie")
	}
	lines := strings.Split(sout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if strings.Index(line, "32 host LOCAL") != -1 {
			fs := strings.Fields(lines[i-1])
			ipstr := strings.TrimSpace(fs[1])
			ip, err := ipToUint32(ipstr)
			if err != nil {
				return "", fmt.Errorf("failed to convert ip, %s: %w", ipstr, err)
			}
			if isIPInRange(ri, ip) {
				return ipstr, nil
			}
		}
	}
	return "", errors.New("unable to find IP")
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
