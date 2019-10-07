package hostsapi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

const hostspath = "C:/Windows/System32/drivers/etc/hosts"

// HostEntry data structure for IP and hostnames
type HostEntry struct {
	idx      int
	IP       string
	Hostname string
}

// HostsAPI data structure
type HostsAPI struct {
	filter    string
	hostsfile *os.File
	entries   map[string]*HostEntry
	remidxs   map[int]interface{}
}

func parseHostfileLine(idx int, line string) ([]*HostEntry, error) {
	if len(line) <= 0 {
		return nil, errors.New("invalid line")
	}
	line = strings.TrimSpace(line)
	if line[0] == '#' {
		return nil, errors.New("comment line")
	}
	fields := strings.Fields(line)
	var validfields []string
	for _, f := range fields {
		if len(f) <= 0 {
			continue
		}
		if f[0] == '#' { // inline comment
			break // don't process any more
		}
		validfields = append(validfields, f)
	}
	if len(validfields) <= 1 {
		return nil, fmt.Errorf("invalid fields for line: %q", line)
	}
	var entries []*HostEntry
	for _, hostname := range validfields[1:] {
		entries = append(entries, &HostEntry{
			idx:      idx,
			IP:       validfields[0],
			Hostname: hostname,
		})
	}

	return entries, nil
}

func (h *HostsAPI) loadAndParse() error {
	scanner := bufio.NewScanner(h.hostsfile)
	idx := 0
	for scanner.Scan() {
		line := scanner.Text()
		entries, err := parseHostfileLine(idx, line)
		idx++
		if err != nil {
			// log.Println(err) // debug
			continue
		}
		for _, e := range entries {
			if h.filter == "" || strings.Contains(e.Hostname, h.filter) {
				h.entries[e.Hostname] = e
				h.remidxs[e.idx] = nil
			}
		}
	}
	h.hostsfile.Seek(0, 0)
	return nil
}

// CreateAPI creates a new instance of the hosts file API
// Call Close() when finished
// `filter` proves ability to filter by string contains
func CreateAPI(filter string) (*HostsAPI, error) {
	f, err := os.Open(hostspath)
	if err != nil {
		return nil, fmt.Errorf("failed to open hosts file: %w", err)
	}
	h := &HostsAPI{
		filter:    filter,
		remidxs:   make(map[int]interface{}),
		entries:   make(map[string]*HostEntry),
		hostsfile: f,
	}
	err = h.loadAndParse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse hosts file: %w", err)
	}
	return h, nil
}

// Close closes the hosts file
func (h *HostsAPI) Close() error {
	err := h.hostsfile.Close()
	if err != nil {
		return fmt.Errorf("failed to close hosts file: %w", err)
	}

	return nil
}

// Entries returns parsed entries of host file
func (h *HostsAPI) Entries() map[string]*HostEntry {
	return h.entries
}

// RemoveEntry removes existing entry from hosts file
func (h *HostsAPI) RemoveEntry(hostname string) error {
	if _, exists := h.entries[hostname]; exists {
		delete(h.entries, hostname)
	} else {
		return fmt.Errorf("failed to remove, hostname does not exist: %s", hostname)
	}
	return nil
}

// AddEntry adds a new HostEntry
func (h *HostsAPI) AddEntry(entry *HostEntry) error {
	if _, exists := h.entries[entry.Hostname]; exists {
		return fmt.Errorf("failed to add entry, hostname already exists: %s", entry.Hostname)
	}

	h.entries[entry.Hostname] = entry

	return nil
}

// Write
func (h *HostsAPI) Write() error {
	var outbuf bytes.Buffer

	// first remove all current entries
	scanner := bufio.NewScanner(h.hostsfile)
	for idx := 0; scanner.Scan() == true; idx++ {
		line := scanner.Text()
		if _, exists := h.remidxs[idx]; !exists {
			outbuf.WriteString(line)
			outbuf.WriteString("\r\n")
		}
	}

	// append entries to file
	for _, e := range h.entries {
		outbuf.WriteString(fmt.Sprintf("%s %s    # managed by wsl2-host\r\n", e.IP, e.Hostname))
	}

	fmt.Println(string(outbuf.Bytes()))
	f, err := os.OpenFile(hostspath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open hosts file for writing: %w", err)
	}
	defer f.Close()

	f.Write(outbuf.Bytes())

	return nil
}
