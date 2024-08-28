package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	knownHome      = os.Getenv("HOME")
	knownConfig    = filepath.Join(knownHome, ".config")
	knownLog       = filepath.Join(knownHome, ".local", "share")
	knownBinary, _ = os.Executable()
)

func TestNew(t *testing.T) {
	p := New()
	assert := assert.New(t)

	assert.Implements((*Config)(nil), p)

	// HOME has to be the same
	got := p.Home()
	assert.Equal(got, knownHome)

	// as long as we're in a subdirectory of the knownConfig, we succeed
	got = p.ConfigDir()
	assert.Contains(got, knownConfig, "ConfigDir matches")

	// binary should be the same
	got = p.Bin()
	assert.Contains(got, knownBinary)

	// as long as we're in a subdirectory of the knownLog, we succeed
	got = p.LogDir()
	assert.Contains(got, knownLog)

	got = p.InstallDir()
	assert.Contains(filepath.Dir(knownBinary), got)
}
