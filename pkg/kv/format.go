package kv

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Format int

const (
	RAW Format = iota + 1
	JSON
	YAML
)

// String implements the Stringer interface
func (e Format) String() string {
	return [...]string{"json", "yaml"}[e-1]
}

// Index makes the index of the Environment retrievable
func (e Format) Index() int {
	return int(e)
}

// IsRAW utility function to quickly determine the format
func (e Format) IsRAW() bool {
	return e == RAW
}

// IsJSON utility function to quickly determine the format
func (e Format) IsJSON() bool {
	return e == JSON
}

// IsYAML utility function to quickly determine the format
func (e Format) IsYAML() bool {
	return e == YAML
}

// FormatFromString matches a string value for a known Format returns the
// corresponding Format. If an invalid Format is given, it returns
// YAML along is a non-nil error
func FormatFromString(s string) (Format, error) {
	switch strings.ToLower(s) {
	case JSON.String():
		return JSON, nil
	case YAML.String():
		return YAML, nil
	case RAW.String():
		return RAW, nil
	default:
		return RAW, fmt.Errorf("invalid file format: %s", s)
	}
}

// Formatter is an object that knows how to format string or byte-slice input
type Formatter interface {
	Format(input string) (string, error)
	FormatBytes(input []byte) ([]byte, error)
}

// JSONFormatter implements the JSON format
type JSONFormatter struct{}

func (jf JSONFormatter) Format(input string) (string, error) {
	js, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func (jf JSONFormatter) FormatBytes(input []byte) ([]byte, error) {
	js, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return js, nil
}

// YAMLFormatter implements the JSON format
type YAMLFormatter struct{}

func (yf YAMLFormatter) Format(input string) (string, error) {
	ya, err := yaml.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(ya), nil
}

func (yf YAMLFormatter) FormatBytes(input []byte) ([]byte, error) {
	ya, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	return ya, nil
}
