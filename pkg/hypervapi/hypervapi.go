package hypervapi

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

const (
	defaultDomainNameSuffix = ".exmaple.com"
	defaultComment          = "managed by api - hyper-vm"
)

type VMManager interface {
	GetRunningVMs() ([]VMInfo, error)
	GetVMIPs(vmName string) ([]string, error)
}

type IP struct {
	IPv4List []string
	IPv6List []string
}

func (ip *IP) String() string {
	if ip == nil {
		return ""
	}
	return fmt.Sprintf("ipv4:%s, ipv6:%s", ip.IPv4List, ip.IPv6List)
}

type VMInfo struct {
	Name   string
	IPInfo *IP
}

func (v *VMInfo) GeDefaulttDomainName() string {
	return v.Name + defaultDomainNameSuffix
}

func (v *VMInfo) GetIP() []string {
	if v.IPInfo == nil {
		return []string{}
	}
	// 预分配足够的容量以避免多次分配
	ips := make([]string, 0, len(v.IPInfo.IPv4List)+len(v.IPInfo.IPv6List))
	ips = append(ips, v.IPInfo.IPv4List...)
	ips = append(ips, v.IPInfo.IPv6List...)
	return ips
}

func (v *VMInfo) GetIPV4() []string {
	if v.IPInfo == nil {
		return []string{}
	}
	// 预分配足够的容量以避免多次分配
	ips := make([]string, 0, len(v.IPInfo.IPv4List))
	ips = append(ips, v.IPInfo.IPv4List...)
	return ips
}

func (v *VMInfo) GetComent() string {
	return defaultComment
}

type HyperVManager struct{}

func NewHyperVManager() *HyperVManager {
	return &HyperVManager{}
}

func (h *HyperVManager) GetRunningVMs() ([]*VMInfo, error) {
	vmNames, err := h.GetRunningVMNames()
	if err != nil {
		return nil, fmt.Errorf("get VM names failed: %w", err)
	}

	vmInfos := make([]*VMInfo, 0, len(vmNames))
	var errs []string

	for _, vmName := range vmNames {
		vmInfo := &VMInfo{Name: vmName}
		vmIPs, err := h.GetVMIPByVMName(vmName)
		if err != nil {
			errs = append(errs, fmt.Sprintf("get IP for VM %q failed: %v", vmName, err))
			continue
		}
		vmInfo.IPInfo = vmIPs
		fmt.Printf("vm:%v, ipInfo:%s\n", vmName, vmIPs)
		vmInfos = append(vmInfos, vmInfo)
	}

	if len(errs) > 0 {
		return vmInfos, fmt.Errorf("errors occurred while getting VM information: %s", strings.Join(errs, "; "))
	}

	return vmInfos, nil
}

func (h *HyperVManager) GetRunningVMNames() ([]string, error) {
	// Implementation to get running VMs using PowerShell commands
	cmd := exec.Command("powershell", "Get-VM | Where-Object {$_.State -eq 'Running'} | Select-Object -ExpandProperty Name")
	var out bytes.Buffer
	var errStr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errStr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Running powershell 'powershell Get-VM | Where-Object {$_.State -eq 'Running'} | Select-Object -ExpandProperty Name' failed with err:%v\n, output:%v\n", errStr.String(), out.String())
		return nil, err
	}
	vms := strings.Split(strings.TrimSpace(out.String()), "\r\n")
	fmt.Println("Get running vm names", vms)
	return vms, nil
}

func (h *HyperVManager) GetVMIPByVMName(vmName string) (*IP, error) {
	// Implementation to get IPs of a VM using PowerShell commands
	cmd := exec.Command("powershell", fmt.Sprintf("Get-VMNetworkAdapter -VMName %s | Select-Object -ExpandProperty IPAddresses", vmName))
	var out bytes.Buffer
	var errStr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errStr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Running powershell 'powershell Get-VMNetworkAdapter -VMName %s | Select-Object -ExpandProperty IPAddresses failed with err:%v\n, output:%v\n", vmName, errStr.String(), out.String())
		return nil, err
	}
	ipInfo := &IP{}
	ipList := strings.Split(strings.TrimSpace(out.String()), "\r\n")
	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if parsedIP := net.ParseIP(ip); parsedIP != nil {
			if parsedIP.To4() != nil {
				ipInfo.IPv4List = append(ipInfo.IPv4List, ip)
			} else {
				ipInfo.IPv6List = append(ipInfo.IPv6List, ip)
			}
		}
	}

	return ipInfo, nil
}
