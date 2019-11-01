// Package wsl2hosts provides helpers for managing alias
// host entries
package wsl2hosts

import (
	"fmt"
	"strings"
)

const prefix = "alias:"
const defaultComment = "managed by wsl2-host"

// IsAlias returns true if given string matches alias pattern
func IsAlias(comment string) bool {
	return strings.HasPrefix(comment, prefix)
}

// DistroName returns the name of the WSL distro the host
// entry is an alias for
func DistroName(comment string) (string, error) {
	if !IsAlias(comment) {
		return "", fmt.Errorf("comment is not alias: %s", comment)
	}

	var name string
	for _, c := range comment[len(prefix):] {
		if c == ';' {
			break
		}
		name += string(c)
	}

	name = strings.TrimSpace(name)
	return name, nil
}

// DistroComment returns hosts file comment alias
// for given distro name
func DistroComment(distroname string) string {
	return fmt.Sprintf("%s %s; %s", prefix, distroname, defaultComment)
}

// DefaultComment returns basic comment for managed host entires
func DefaultComment() string {
	return defaultComment
}
