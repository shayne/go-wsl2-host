package wsl2hosts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAlias(t *testing.T) {
	assert.True(t, IsAlias("alias: Ubuntu-18.04; managed by wsl2-host"))
	assert.False(t, IsAlias("managed by wsl2-host"))
}

func TestDistroName(t *testing.T) {
	name, err := DistroName("alias: Ubuntu-18.04; managed by wsl2-host")
	assert.Nil(t, err)
	assert.Equal(t, "Ubuntu-18.04", name)
	name, err = DistroName("alias: ClearLinux; managed by wsl2-host")
	assert.Nil(t, err)
	assert.Equal(t, "ClearLinux", name)
	name, err = DistroName("alias: Foo Bar; managed by wsl2-host")
	assert.Nil(t, err)
	assert.Equal(t, "Foo Bar", name)
	name, err = DistroName("managed by wsl2-host")
	assert.NotNil(t, err)
	assert.Equal(t, "", name)
}

func TestDistroComment(t *testing.T) {
	comment := DistroComment("Ubuntu-18.04")
	assert.Equal(t, "alias: Ubuntu-18.04; managed by wsl2-host", comment)
}

func TestDefaultComment(t *testing.T) {
	assert.Equal(t, "managed by wsl2-host", DefaultComment())
}
