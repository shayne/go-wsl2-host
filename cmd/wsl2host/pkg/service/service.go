package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shayne/go-wsl2-host/pkg/wslcli"
)

const wslHostname = "wsl.local"
const windowsHostname = "windows.local"

// IsRunning returns whether or not WSL is running
func IsRunning() (bool, error) {
	running, err := wslcli.Running()
	if err != nil {
		return false, err
	}
	return running, nil
}

func getWSLIP() (string, error) {
	ip, err := wslcli.GetWSLIP()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(ip), nil
}

// UpdateIP updates the Windows hosts file
func UpdateIP() error {
	wslIP, err := getWSLIP()
	if err != nil {
		return err
	}
	hostIP, err := wslcli.GetHostIP()
	if err != nil {
		return err
	}
	f, err := os.OpenFile("c:/Windows/System32/drivers/etc/hosts", os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	wslExisted := false
	wslWasCorrect := false
	hostExisted := false
	hostWasCorrect := false
	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, 50)

	wslLine := fmt.Sprintf("%s %s", wslIP, wslHostname)
	hostLine := fmt.Sprintf("%s %s", hostIP, windowsHostname)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, wslHostname) {
			if strings.Contains(line, wslIP) {
				wslWasCorrect = true
				lines = append(lines, line)
			} else {
				wslExisted = true
				lines = append(lines, wslLine)
			}
		} else if strings.HasSuffix(line, windowsHostname) {
			if strings.Contains(line, hostIP) {
				hostWasCorrect = true
				lines = append(lines, line)
			} else {
				hostExisted = true
				lines = append(lines, hostLine)
			}
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if !wslWasCorrect && !wslExisted {
		lines = append(lines, wslLine)
	}
	if !hostWasCorrect && !hostExisted {
		lines = append(lines, hostLine)
	}

	_, err = f.WriteAt([]byte(strings.Join(lines, "\r\n")), 0)
	if err != nil {
		return err
	}
	return nil
}
