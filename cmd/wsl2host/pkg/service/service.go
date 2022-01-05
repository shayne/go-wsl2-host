package service

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shayne/go-wsl2-host/internal/wsl2hosts"
	"golang.org/x/sys/windows/svc/debug"

	"github.com/shayne/go-wsl2-host/pkg/hostsapi"

	"github.com/shayne/go-wsl2-host/pkg/wslapi"
)

const tld = ".wsl"
const windowshost = "windows.local"

var hostnamereg, _ = regexp.Compile("[^A-Za-z0-9]+")

func distroNameToHostname(distroname string) string {
	// Ubuntu-18.04
	// => ubuntu1804.wsl
	hostname := strings.ToLower(distroname)
	hostname = hostnamereg.ReplaceAllString(hostname, "")
	return hostname + tld
}

// Run main entry point to service logic
func Run(elog debug.Log) error {
	// Then get all wsl info. and run them with config.
	infos, err := wslapi.GetAllInfo()
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to get infos: %v", err))
		return fmt.Errorf("failed to get infos: %w", err)
	}

	err = updateHostIP(elog, infos)
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to update host IP info: %s", err))
	}

	for _, i := range infos {
		if i.Running {
			err = updateDistroIP(elog, infos, i.Name)
			if err != nil {
				elog.Error(1, fmt.Sprintf("failed to update distro[%s] IP info: %s", i.Name, err))
			}
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func updateHostIP(elog debug.Log, distros []*wslapi.DistroInfo) error {
	// update the ip to the wsl
	hapi, err := hostsapi.CreateAPI("wsl2-host") // filtere only managed host entries
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to create hosts api: %v", err))
		return fmt.Errorf("failed to create hosts api: %w", err)
	}

	updated := false
	hostentries := hapi.Entries()

	// update the wsl ip to host
	for _, i := range distros {
		hostname := distroNameToHostname(i.Name)
		// remove stopped distros
		if !i.Running {
			err := hapi.RemoveEntry(hostname)
			if err == nil {
				updated = true
			}
			continue
		}

		// update IPs of running distros
		if he, exists := hostentries[hostname]; exists {
			if he.IP != i.IP {
				updated = true
				he.IP = i.IP
			}
		} else {
			// add running distros not present
			err := hapi.AddEntry(&hostsapi.HostEntry{
				Hostname: hostname,
				IP:       i.IP,
				Comment:  wsl2hosts.DefaultComment(),
			})
			if err == nil {
				updated = true
			}
		}
	}

	// process aliases
	defdistro, _ := wslapi.GetDefaultDistro()
	if err != nil {
		elog.Error(1, fmt.Sprintf("GetDefaultDistro failed: %v", err))
		return fmt.Errorf("GetDefaultDistro failed: %w", err)
	}
	var aliasmap = make(map[string]interface{})
	defdistroip, _ := wslapi.GetIP(defdistro.Name)
	if defdistro.Running {
		aliases, err := wslapi.GetHostAliases()
		if err == nil {
			for _, a := range aliases {
				aliasmap[a] = nil
			}
		}
	}
	// update entries after distro processing
	hostentries = hapi.Entries()
	for _, he := range hostentries {
		if !wsl2hosts.IsAlias(he.Comment) {
			continue
		}
		// update IP for aliases when running and if it exists in aliasmap
		if _, ok := aliasmap[he.Hostname]; ok && defdistro.Running {
			if he.IP != defdistroip {
				updated = true
				he.IP = defdistroip
			}
		} else { // remove entry when not running or not in aliasmap
			err := hapi.RemoveEntry(he.Hostname)
			if err == nil {
				updated = true
			}
		}
	}

	for hostname := range aliasmap {
		// add new aliases
		if _, ok := hostentries[hostname]; !ok && defdistro.Running {
			err := hapi.AddEntry(&hostsapi.HostEntry{
				IP:       defdistroip,
				Hostname: hostname,
				Comment:  wsl2hosts.DistroComment(defdistro.Name),
			})
			if err == nil {
				updated = true
			}
		}
	}

	hostIP, err := hostsapi.GetHostIP()

	if err == nil {
		hostname, err := os.Hostname()
		hostAlias := distroNameToHostname(hostname)
		err = hapi.AddEntry(&hostsapi.HostEntry{
			IP:       hostIP,
			Hostname: hostAlias,
			Comment:  wsl2hosts.DistroComment(hostname),
		})

		if err == nil {
			updated = true
		}
	}

	if updated {
		err = hapi.Write()
		if err != nil {
			elog.Error(1, fmt.Sprintf("failed to write hosts file: %v", err))
			return fmt.Errorf("failed to write hosts file: %w", err)
		}

		// restart the IP Helper service (iphlpsvc) for port forwarding
		exec.Command("C:\\Windows\\System32\\cmd.exe", "/C net stop  iphlpsvc").Run()
		exec.Command("C:\\Windows\\System32\\cmd.exe", "/C net start iphlpsvc").Run()
	}

	return nil
}

/// Write all other distro and host into the hosts file for each distro.
func updateDistroIP(elog debug.Log, distros []*wslapi.DistroInfo, distro string) error {
	host_ip, err := hostsapi.GetHostIP()
	if err != nil {
		return err
	}
	err = wslapi.AddOrUpdateHostIP(distro, windowshost, host_ip)
	if err != nil {
		return err
	}

	for _, dist := range distros {
		if !dist.Running || dist.Name == distro {
			continue
		}
		hostAlias := distroNameToHostname(dist.Name)
		err = wslapi.AddOrUpdateHostIP(distro, hostAlias, dist.IP)
		if err != nil {
			return err
		}
	}
	return nil
}
