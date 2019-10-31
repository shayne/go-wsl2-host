// +build windows

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

// InstallService installs the Windows service and starts it
func InstallService(name, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}
	fmt.Printf("Windows Username: ")
	var username string
	fmt.Scanln(&username)
	if !strings.Contains(username, "\\") && !strings.Contains(username, "@") {
		username = fmt.Sprintf(".\\%s", strings.TrimSpace(username))
	}
	fmt.Printf("Windows Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	password := strings.TrimSpace(string(bytePassword))
	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: desc, StartType: mgr.StartAutomatic, ServiceStartName: username, Password: password}, "is", "auto-started")
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	StartService(name)
	return nil
}

// RemoveService uninstalls the Windows service
func RemoveService(name string) error {
	var err2 error
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		err2 = fmt.Errorf("service %s is not installed", name)
	} else {
		defer s.Close()
		err = s.Delete()
		if err != nil {
			err2 = fmt.Errorf("%w; %s", err2, err.Error())
		}
	}
	err = eventlog.Remove(name)
	if err != nil {
		err2 = fmt.Errorf("%w; RemoveEventLogSource() failed: %s", err2, err)
	}
	if err2 != nil {
		return err2
	}
	return nil
}
