package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	knownHome      = os.Getenv("HOME")
	knownConfig    = filepath.Join(knownHome, ".config")
	knownDataOrLog = filepath.Join(knownHome, ".local", "share")
)

func TestNew(t *testing.T) {
	p, _ := Paths()
	asrt := assert.New(t)

	// as long as we're in a subdirectory of the knownConfig, we succeed
	got := p.Config
	asrt.Contains(got, knownConfig, "ConfigDir matches")

	// test data path
	got = p.Data
	asrt.Contains(got, knownDataOrLog)

	// test log path
	got = p.Log
	asrt.Contains(got, knownDataOrLog)

	// test cache path
	osc, _ := os.UserCacheDir()
	got = p.Cache
	asrt.Contains(got, osc)
}
