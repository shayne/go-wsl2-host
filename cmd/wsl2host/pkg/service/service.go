package service

import (
	"fmt"
	"regexp"
	"strings"
	"os/exec"

	"github.com/shayne/go-wsl2-host/internal/wsl2hosts"
	"golang.org/x/sys/windows/svc/debug"

	"github.com/shayne/go-wsl2-host/pkg/hostsapi"

	"github.com/shayne/go-wsl2-host/pkg/wslapi"
)

const tld = ".wsl"

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
	infos, err := wslapi.GetAllInfo()
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to get infos: %v", err))
		return fmt.Errorf("failed to get infos: %w", err)
	}

	hapi, err := hostsapi.CreateAPI("wsl2-host") // filtere only managed host entries
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to create hosts api: %v", err))
		return fmt.Errorf("failed to create hosts api: %w", err)
	}

	updated := false
	hostentries := hapi.Entries()

	for _, i := range infos {
		hostname := distroNameToHostname(i.Name)
		// remove stopped distros
		if i.Running == false {
			err := hapi.RemoveEntry(hostname)
			if err == nil {
				updated = true
			}
			continue
		}

		// update IPs of running distros
		var ip string
		if i.Version == 1 {
			ip = "127.0.0.1"
		} else {
			ip, err = wslapi.GetIP(i.Name)
			if err != nil {
				elog.Info(1, fmt.Sprintf("failed to get IP for distro %q: %v", i.Name, err))
				continue
			}
		}
		if he, exists := hostentries[hostname]; exists {
			if he.IP != ip {
				updated = true
				he.IP = ip
			}
		} else {
			// add running distros not present
			err := hapi.AddEntry(&hostsapi.HostEntry{
				Hostname: hostname,
				IP:       ip,
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
