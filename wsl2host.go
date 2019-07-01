package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shayne/go-wsl2-host/pkg/wslcli"
)

const hostname = "wsl.local"

func isRunning() (bool, error) {
	running, err := wslcli.Running()
	if err != nil {
		return false, err
	}
	return running, nil
}

func getIP() (string, error) {
	ip, err := wslcli.IP()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(ip), nil
}

func updateIP() error {
	ip, err := getIP()
	if err != nil {
		return err
	}
	f, err := os.OpenFile("c:/Windows/System32/drivers/etc/hosts", os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	n := 0
	wslexisting := false
	scanner := bufio.NewScanner(f)
	lines := make([]string, 0, 50)

	wslline := fmt.Sprintf("%s %s", ip, hostname)

	for scanner.Scan() {
		n++
		line := scanner.Text()
		if strings.HasSuffix(line, hostname) {
			if strings.Contains(line, ip) {
				return nil
			}
			wslexisting = true
			lines = append(lines, wslline)
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if !wslexisting {
		lines = append(lines, wslline)
	}

	lines = append(lines, "\r\n\r\n")

	_, err = f.WriteAt([]byte(strings.Join(lines, "\r\n")), 0)
	if err != nil {
		return err
	}
	return nil
}
